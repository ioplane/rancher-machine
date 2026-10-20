package rpcdriver

import (
	"net/rpc"

	"github.com/rancher/machine/libmachine/drivers"
)

const prepareCreateRecoveryMethod = ".PrepareCreateRecovery"

// CreateRecoveryRequest carries current flags without loading saved credentials.
type CreateRecoveryRequest struct {
	Flags   drivers.DriverOptions
	Context drivers.CreateRecoveryContext
}

func (c *RPCClientDriver) PrepareCreateRecovery(flags drivers.DriverOptions, context drivers.CreateRecoveryContext) (bool, error) {
	var recovering bool
	err := c.Client.Call(prepareCreateRecoveryMethod, CreateRecoveryRequest{Flags: flags, Context: context}, &recovering)
	if serverError, ok := err.(rpc.ServerError); ok && string(serverError) == "rpc: can't find method "+c.Client.rpcServiceName+prepareCreateRecoveryMethod {
		return false, nil
	}
	return recovering, err
}

func (r *RPCServerDriver) PrepareCreateRecovery(request CreateRecoveryRequest, reply *bool) error {
	recovery, ok := r.ActualDriver.(drivers.CreateRecoveryDriver)
	if !ok {
		*reply = false
		return nil
	}
	var err error
	*reply, err = recovery.PrepareCreateRecovery(request.Flags, request.Context)
	return err
}

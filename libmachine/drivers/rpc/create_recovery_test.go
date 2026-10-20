package rpcdriver

import (
	"errors"
	"net"
	"net/rpc"
	"testing"

	"github.com/rancher/machine/drivers/fakedriver"
	"github.com/rancher/machine/libmachine/drivers"
)

type legacyRecoveryRPC struct{}

func (*legacyRecoveryRPC) Existing(_ struct{}, reply *bool) error { *reply = true; return nil }

type failingRecoveryRPC struct{}

func (*failingRecoveryRPC) PrepareCreateRecovery(_ CreateRecoveryRequest, _ *bool) error {
	return errors.New("binding mismatch")
}

type capableRecoveryDriver struct{ *fakedriver.Driver }

func (*capableRecoveryDriver) PrepareCreateRecovery(opts drivers.DriverOptions, c drivers.CreateRecoveryContext) (bool, error) {
	return opts.String("current") == "yes" && c.Existing && c.HostnameOverride == "chosen", nil
}

func recoveryRPCClient(t *testing.T, service interface{}) *RPCClientDriver {
	t.Helper()
	server := rpc.NewServer()
	if err := server.RegisterName(RPCServiceNameV1, service); err != nil {
		t.Fatal(err)
	}
	left, right := net.Pipe()
	go server.ServeConn(left)
	client := rpc.NewClient(right)
	t.Cleanup(func() { _ = client.Close() })
	return &RPCClientDriver{Client: NewInternalClient(client)}
}

func TestPrepareCreateRecoveryRPC(t *testing.T) {
	for _, tc := range []struct {
		name    string
		service interface{}
		want    bool
		wantErr bool
	}{
		{name: "legacy missing method", service: &legacyRecoveryRPC{}},
		{name: "unsupported driver", service: NewRPCServerDriver(&fakedriver.Driver{})},
		{name: "supported driver", service: NewRPCServerDriver(&capableRecoveryDriver{}), want: true},
		{name: "driver error", service: &failingRecoveryRPC{}, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := recoveryRPCClient(t, tc.service)
			got, err := c.PrepareCreateRecovery(&RPCFlags{Values: map[string]interface{}{"current": "yes"}}, drivers.CreateRecoveryContext{Existing: true, HostnameOverride: "chosen"})
			if got != tc.want || (err != nil) != tc.wantErr {
				t.Fatalf("got=%v err=%v", got, err)
			}
		})
	}
	t.Run("transport failure", func(t *testing.T) {
		c := recoveryRPCClient(t, &legacyRecoveryRPC{})
		_ = c.Client.RPCClient.Close()
		got, err := c.PrepareCreateRecovery(&RPCFlags{}, drivers.CreateRecoveryContext{})
		if got || err == nil {
			t.Fatalf("got=%v err=%v", got, err)
		}
	})
}

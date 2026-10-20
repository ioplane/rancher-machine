package drivers

// CreateRecoveryContext identifies the current SSH custom-script create request.
// It is optional and does not change the Driver contract or RPC protocol version.
type CreateRecoveryContext struct {
	Existing                bool
	CustomInstallScriptPath string
	HostnameOverride        string
}

// CreateRecoveryDriver may adopt a proved pending create using current flags.
// A true reply means flags are already configured; it never authorizes replay.
type CreateRecoveryDriver interface {
	PrepareCreateRecovery(DriverOptions, CreateRecoveryContext) (bool, error)
}

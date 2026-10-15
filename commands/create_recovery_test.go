package commands

import (
	"errors"
	"testing"

	"github.com/rancher/machine/commands/commandstest"
	"github.com/rancher/machine/drivers/fakedriver"
	"github.com/rancher/machine/libmachine/drivers"
	"github.com/rancher/machine/libmachine/host"
	"github.com/rancher/machine/libmachine/libmachinetest"
	"github.com/rancher/machine/libmachine/mcnerror"
	"github.com/rancher/machine/libmachine/mcnflag"
)

type recoveryCommandLine struct{ *commandstest.FakeCommandLine }

func (*recoveryCommandLine) FlagNames() []string { return nil }

type recoveryCommandDriver struct {
	*fakedriver.Driver
	eligible          bool
	prepareErr        error
	prepares, configs int
	context           drivers.CreateRecoveryContext
}

func (d *recoveryCommandDriver) GetCreateFlags() []mcnflag.Flag { return nil }
func (d *recoveryCommandDriver) SetConfigFromFlags(drivers.DriverOptions) error {
	d.configs++
	return nil
}

func (d *recoveryCommandDriver) PrepareCreateRecovery(_ drivers.DriverOptions, c drivers.CreateRecoveryContext) (bool, error) {
	d.prepares++
	d.context = c
	return d.eligible, d.prepareErr
}

type recoveryCommandAPI struct {
	libmachinetest.FakeAPI
	current               *host.Host
	exists                bool
	creates, saves, loads int
}

func (a *recoveryCommandAPI) NewHost(string, []byte) (*host.Host, error) { return a.current, nil }
func (a *recoveryCommandAPI) Exists(string) (bool, error)                { return a.exists, nil }
func (a *recoveryCommandAPI) Create(h *host.Host) error {
	if h != a.current {
		panic("stale host")
	}
	a.creates++
	return nil
}
func (a *recoveryCommandAPI) Save(*host.Host) error { a.saves++; return nil }
func (a *recoveryCommandAPI) Load(string) (*host.Host, error) {
	a.loads++
	panic("recovery must use incoming host")
}

func TestCreateRecoveryUsesCurrentHost(t *testing.T) {
	for _, tc := range []struct {
		name               string
		eligible, existing bool
		prepareErr         error
		wantCreate         int
	}{
		{name: "pending", eligible: true, existing: true, wantCreate: 1},
		{name: "legacy", existing: true},
		{name: "complete", existing: true},
		{name: "removed", existing: true},
		{name: "mismatch", eligible: true, existing: true, prepareErr: errors.New("binding mismatch")},
		{name: "fresh", wantCreate: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := &recoveryCommandDriver{Driver: &fakedriver.Driver{}, eligible: tc.eligible, prepareErr: tc.prepareErr}
			api := &recoveryCommandAPI{current: &host.Host{Name: "node", Driver: d, DriverName: "nutanix"}, exists: tc.existing}
			c := &commandstest.FakeCommandLine{CliArgs: []string{"node"}, LocalFlags: &commandstest.FakeFlagger{Data: map[string]interface{}{"driver": "nutanix", "custom-install-script": "/absent-script", "hostname-override": "chosen"}}, GlobalFlags: &commandstest.FakeFlagger{Data: map[string]interface{}{}}}
			err := cmdCreate(&recoveryCommandLine{c}, api)
			if tc.prepareErr != nil {
				if err == nil {
					t.Fatal("missing refusal")
				}
			} else if tc.existing && !tc.eligible {
				var exists mcnerror.ErrHostAlreadyExists
				if !errors.As(err, &exists) {
					t.Fatalf("expected existing host, got %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if api.creates != tc.wantCreate || api.saves != tc.wantCreate || api.loads != 0 {
				t.Fatalf("creates=%d saves=%d loads=%d", api.creates, api.saves, api.loads)
			}
			if d.prepares != 1 || d.context.Existing != tc.existing || d.context.CustomInstallScriptPath != "/absent-script" || d.context.HostnameOverride != "chosen" {
				t.Fatalf("prepare/context=%d %+v", d.prepares, d.context)
			}
			wantConfigs := tc.wantCreate
			if tc.eligible {
				wantConfigs = 0
			}
			if d.configs != wantConfigs {
				t.Fatalf("configs=%d want=%d", d.configs, wantConfigs)
			}
		})
	}
}

func TestExistingUnsupportedHostKeepsOriginalError(t *testing.T) {
	api := &recoveryCommandAPI{current: &host.Host{Name: "node", Driver: &fakedriver.Driver{}, DriverName: "fake"}, exists: true}
	c := &commandstest.FakeCommandLine{CliArgs: []string{"node"}, LocalFlags: &commandstest.FakeFlagger{Data: map[string]interface{}{"driver": "fake", "custom-install-script": "/absent-script"}}, GlobalFlags: &commandstest.FakeFlagger{Data: map[string]interface{}{}}}
	err := cmdCreate(&recoveryCommandLine{c}, api)
	var exists mcnerror.ErrHostAlreadyExists
	if !errors.As(err, &exists) || api.creates != 0 || api.saves != 0 || api.loads != 0 {
		t.Fatalf("err=%v creates=%d saves=%d loads=%d", err, api.creates, api.saves, api.loads)
	}
}

type userdataRecoveryDriver struct{ *recoveryCommandDriver }

func (*userdataRecoveryDriver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{mcnflag.StringFlag{Name: "user-data"}}
}

func TestExistingUserdataHostKeepsOriginalError(t *testing.T) {
	d := &recoveryCommandDriver{Driver: &fakedriver.Driver{}, eligible: true}
	api := &recoveryCommandAPI{current: &host.Host{Name: "node", Driver: &userdataRecoveryDriver{d}}, exists: true}
	c := &commandstest.FakeCommandLine{CliArgs: []string{"node"}, LocalFlags: &commandstest.FakeFlagger{Data: map[string]interface{}{"custom-install-script": "/absent-script"}}, GlobalFlags: &commandstest.FakeFlagger{Data: map[string]interface{}{}}}
	err := cmdCreate(&recoveryCommandLine{c}, api)
	var exists mcnerror.ErrHostAlreadyExists
	if !errors.As(err, &exists) || d.prepares != 0 || d.configs != 0 || api.creates != 0 || api.saves != 0 {
		t.Fatalf("err=%v prepares=%d configs=%d creates=%d saves=%d", err, d.prepares, d.configs, api.creates, api.saves)
	}
}

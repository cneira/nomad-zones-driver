package zone

import (
	"context"
	"fmt"
	"time"

	zconfig "git.wegmueller.it/illumos/go-zone/config"
	"git.wegmueller.it/illumos/go-zone/lifecycle"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/drivers/shared/eventer"
	"github.com/hashicorp/nomad/plugins/base"
	"github.com/hashicorp/nomad/plugins/drivers"
	"github.com/hashicorp/nomad/plugins/shared/hclspec"
	pstructs "github.com/hashicorp/nomad/plugins/shared/structs"
)

const (
	// pluginName is the name of the plugin
	pluginName = "zone"

	// fingerprintPeriod is the interval at which the driver will send fingerprint responses
	fingerprintPeriod = 30 * time.Second

	// taskHandleVersion is the version of task handle which this driver sets
	// and understands how to decode driver state
	taskHandleVersion = 1
)

var (
	// pluginInfo is the response returned for the PluginInfo RPC
	pluginInfo = &base.PluginInfoResponse{
		Type:              base.PluginTypeDriver,
		PluginApiVersions: []string{drivers.ApiVersion010},
		PluginVersion:     "0.1.1-dev",
		Name:              pluginName,
	}

	// taskConfigSpec is the hcl specification for the driver config section of
	// a task within a job. It is returned in the TaskConfigSchema RPC
	taskConfigSpec = hclspec.NewObject(map[string]*hclspec.Spec{
		"Zonepath":        hclspec.NewAttr("Zonepath", "string", true),
		"HostId":          hclspec.NewAttr("HostId", "string", false),
		"Autoboot":        hclspec.NewAttr("Autoboot", "string", false),
		"SchedulingClass": hclspec.NewAttr("SchedulingClass", "string", false),
		"Brand":           hclspec.NewAttr("Brand", "string", false),
		"CpuShares":       hclspec.NewAttr("CpuShares", "string", false),
		"DedicatedCpu":    hclspec.NewAttr("DedicatedCpu", "string", false),
		"CappedMemory":    hclspec.NewAttr("CappedMemory", "string", false),
		"Docker":          hclspec.NewAttr("Docker", "string", false),
		"LockedMemory":    hclspec.NewAttr("LockedMemory", "string", false),
		"SwapMemory":      hclspec.NewAttr("SwapMemory", "string", false),
		"ShmMemory":       hclspec.NewAttr("ShmMemory", "string", false),
		"SemIds":          hclspec.NewAttr("SemIds", "string", false),
		"MsgIds":          hclspec.NewAttr("MsgIds", "string", false),
		"ShmIds":          hclspec.NewAttr("ShmIds", "string", false),
		"Lwps":            hclspec.NewAttr("Lwps", "string", false),
		"Attributes": hclspec.NewBlockList("Attributes", hclspec.NewObject(map[string]*hclspec.Spec{
			"Name":  hclspec.NewAttr("Name", "string", false),
			"Type":  hclspec.NewAttr("Type", "string", false),
			"Value": hclspec.NewAttr("Value", "string", false),
		})),
		"FileSystems": hclspec.NewBlockList("FileSystems", hclspec.NewObject(map[string]*hclspec.Spec{
			"Dir":     hclspec.NewAttr("Dir", "string", false),
			"Special": hclspec.NewAttr("Special", "string", false),
			"Type":    hclspec.NewAttr("Type", "string", false),
			"raw":     hclspec.NewAttr("raw", "string", false),
			"Fsoption": hclspec.NewBlockList("Fsoption", hclspec.NewObject(map[string]*hclspec.Spec{
				"Name": hclspec.NewAttr("Name", "string", false)})),
		})),
		"Devices": hclspec.NewBlockList("Devices", hclspec.NewObject(map[string]*hclspec.Spec{
			"Match": hclspec.NewAttr("Match", "string", false),
		})),
		"IpType": hclspec.NewAttr("IpType", "string", false),
		"Networks": hclspec.NewBlockList("Networks", hclspec.NewObject(map[string]*hclspec.Spec{
			"Address":        hclspec.NewAttr("Address", "string", false),
			"Physical":       hclspec.NewAttr("Physical", "string", false),
			"Defrouter":      hclspec.NewAttr("Defrouter", "string", false),
			"AllowedAddress": hclspec.NewAttr("AllowedAddress", "string", false),
		})),
	})

	// capabilities is returned by the Capabilities RPC and indicates what
	// optional features this driver supports
	capabilities = &drivers.Capabilities{
		SendSignals: false,
		Exec:        false,
		FSIsolation: drivers.FSIsolationImage,
	}
)

type Driver struct {
	// eventer is used to handle multiplexing of TaskEvents calls such that an
	// event can be broadcast to all callers
	eventer *eventer.Eventer

	// config is the driver configuration set by the SetConfig RPC
	config *Config

	// nomadConfig is the client config from nomad
	nomadConfig *base.ClientDriverConfig

	// tasks is the in memory datastore mapping taskIDs to rawExecDriverHandles
	tasks *taskStore

	// ctx is the context for the driver. It is passed to other subsystems to
	// coordinate shutdown
	ctx context.Context

	// signalShutdown is called when the driver is shutting down and cancels the
	// ctx passed to any subsystems
	signalShutdown context.CancelFunc

	// logger will log to the Nomad agent
	logger hclog.Logger
}

// Config is the driver configuration set by the SetConfig RPC call
type Config struct {
}

// TaskConfig is the driver configuration of a task within a job
type TaskConfig struct {
	Zonepath        string               `codec:"Zonepath"`
	HostId          string               `code:"HostId"`
	Brand           string               `codec:"Brand"`
	Docker          string               `code:"Docker"`
	Autoboot        string               `codec:"Autoboot"`
	SchedulingClass string               `code:"SchedulingClass"`
	CpuShares       string               `codec:"CpuShares"`
	CappedMemory    string               `codec:"CappedMemory"`
	LockedMemory    string               `codec:"LockedMemory"`
	SwapMemory      string               `code:"SwapMemory"`
	ShmMemory       string               `code:"ShmMemory"`
	DedicatedCpu    string               `code:"DedicatedCpu"`
	SemIds          string               `code:"SemIds"`
	ShmIds          string               `code:"ShmIds"`
	MsgIds          string               `code:"MsgIds"`
	Lwps            string               `codec:"Lwps"`
	IpType          string               `code:"IpType"`
	Networks        []zconfig.Network    `codec:"Networks"`
	Attributes      []zconfig.Attribute  `code:"Attributes"`
	FileSystems     []zconfig.FileSystem `code:"FileSystems"`
	Devices         []zconfig.Device     `code:"Devices"`
}

// TaskState is the state which is encoded in the handle returned in
// StartTask. This information is needed to rebuild the task state and handler
// during recovery.
type TaskState struct {
	TaskConfig    *drivers.TaskConfig
	ContainerName string
	StartedAt     time.Time
}

func NewZoneDriver(logger hclog.Logger) drivers.DriverPlugin {
	ctx, cancel := context.WithCancel(context.Background())
	logger = logger.Named(pluginName)
	return &Driver{
		eventer:        eventer.NewEventer(ctx, logger),
		config:         &Config{},
		tasks:          newTaskStore(),
		ctx:            ctx,
		signalShutdown: cancel,
		logger:         logger,
	}
}

func (d *Driver) PluginInfo() (*base.PluginInfoResponse, error) {
	return pluginInfo, nil
}

func (d *Driver) ConfigSchema() (*hclspec.Spec, error) {
	return nil, nil
}

func (d *Driver) SetConfig(cfg *base.Config) error {
	var config Config
	if len(cfg.PluginConfig) != 0 {
		if err := base.MsgPackDecode(cfg.PluginConfig, &config); err != nil {
			return err
		}
	}

	d.config = &config
	if cfg.AgentConfig != nil {
		d.nomadConfig = cfg.AgentConfig.Driver
	}

	return nil
}

func (d *Driver) Shutdown(ctx context.Context) error {
	d.signalShutdown()
	return nil
}

func (d *Driver) TaskConfigSchema() (*hclspec.Spec, error) {
	return taskConfigSpec, nil
}

func (d *Driver) Capabilities() (*drivers.Capabilities, error) {
	return capabilities, nil
}

func (d *Driver) Fingerprint(ctx context.Context) (<-chan *drivers.Fingerprint, error) {
	ch := make(chan *drivers.Fingerprint)
	go d.handleFingerprint(ctx, ch)
	return ch, nil
}

func (d *Driver) handleFingerprint(ctx context.Context, ch chan<- *drivers.Fingerprint) {
	defer close(ch)
	ticker := time.NewTimer(0)
	for {
		select {
		case <-ctx.Done():
			return
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			ticker.Reset(fingerprintPeriod)
			ch <- d.buildFingerprint()
		}
	}
}

func (d *Driver) buildFingerprint() *drivers.Fingerprint {
	var health drivers.HealthState
	var desc string
	attrs := map[string]*pstructs.Attribute{"driver.zone": pstructs.NewStringAttribute("1")}
	health = drivers.HealthStateHealthy
	desc = "ready"
	d.logger.Info("buildFingerprint()", "driver.FingerPrint", hclog.Fmt("%+v", health))
	return &drivers.Fingerprint{
		Attributes:        attrs,
		Health:            health,
		HealthDescription: desc,
	}
}

func (d *Driver) RecoverTask(handle *drivers.TaskHandle) error {
	if handle == nil {
		return fmt.Errorf("error: handle cannot be nil")
	}

	if _, ok := d.tasks.Get(handle.Config.ID); ok {
		return nil
	}

	var taskState TaskState
	if err := handle.GetDriverState(&taskState); err != nil {
		return fmt.Errorf("failed to decode task state from handle: %v", err)
	}

	z := zconfig.New(taskState.ContainerName)

	mgr, err := lifecycle.NewManager(z)
	if err != nil {
		return fmt.Errorf("Cannot create mgr %v", err)
	}

	if err = mgr.Reboot(nil); err != nil {
		return fmt.Errorf("Cannot Reboot zone err= %+v", err)
	}

	h := &taskHandle{
		container:  zconfig.Zone{Brand: z.Brand, Zonepath: z.Zonepath},
		taskConfig: taskState.TaskConfig,
		State:      drivers.TaskStateRunning,
		startedAt:  taskState.StartedAt,
		exitResult: &drivers.ExitResult{},
		logger:     d.logger,
	}

	d.tasks.Set(taskState.TaskConfig.ID, h)
	go h.run()
	return nil
}

func (d *Driver) StartTask(cfg *drivers.TaskConfig) (*drivers.TaskHandle, *drivers.DriverNetwork, error) {

	if _, ok := d.tasks.Get(cfg.ID); ok {
		return nil, nil, fmt.Errorf("task with ID %q already started", cfg.ID)
	}

	var driverConfig TaskConfig
	if err := cfg.DecodeDriverConfig(&driverConfig); err != nil {
		return nil, nil, fmt.Errorf("failed to decode driver config: %v", err)
	}

	d.logger.Info("starting zone task", "driver_cfg", hclog.Fmt("%+v", driverConfig))
	handle := drivers.NewTaskHandle(taskHandleVersion)
	handle.Config = cfg

	z := d.initializeContainer(cfg, driverConfig)
	if err := z.WriteToFile(); err != nil {
		return nil, nil, fmt.Errorf("Cannot write file %q", cfg.ID)
	}
	if err := zconfig.Register(z); err != nil {
		zconfig.Unregister(z)
		z.RemoveFile()
		return nil, nil, fmt.Errorf("Cannot Register %q, err=%+v", cfg.ID, err)
	}
	mgr, err := lifecycle.NewManager(z)
	if err != nil {
		return nil, nil, fmt.Errorf("Cannot create mgr %q", cfg.ID)
	}

	if err = mgr.Verify(); err != nil {
		return nil, nil, fmt.Errorf("Error Verifying zone  %q, err= %+v", cfg.ID, err)
	}

	if err = mgr.Install(nil); err != nil {
		return nil, nil, fmt.Errorf("Cannot install zone %q, err= %+v", cfg.ID, err)
	}

	if err = mgr.Boot(nil); err != nil {
		return nil, nil, fmt.Errorf("Cannot boot zone %q, err= %+v", cfg.ID, err)
	}

	h := &taskHandle{
		container:  zconfig.Zone{Brand: z.Brand, Zonepath: z.Zonepath},
		taskConfig: cfg,
		State:      drivers.TaskStateRunning,
		startedAt:  time.Now().Round(time.Millisecond),
		logger:     d.logger,
	}

	driverState := TaskState{
		ContainerName: z.Name,
		TaskConfig:    cfg,
		StartedAt:     h.startedAt,
	}

	if err := handle.SetDriverState(&driverState); err != nil {
		d.logger.Error("failed to start task, error setting driver state", "error", err)
		return nil, nil, fmt.Errorf("failed to set driver state: %v", err)
	}

	d.tasks.Set(cfg.ID, h)

	go h.run()

	return handle, nil, nil
}

func (d *Driver) WaitTask(ctx context.Context, taskID string) (<-chan *drivers.ExitResult, error) {
	handle, ok := d.tasks.Get(taskID)
	if !ok {
		return nil, drivers.ErrTaskNotFound
	}

	ch := make(chan *drivers.ExitResult)
	go d.handleWait(ctx, handle, ch)

	return ch, nil
}

func (d *Driver) handleWait(ctx context.Context, handle *taskHandle, ch chan *drivers.ExitResult) {
	defer close(ch)

	// Going with simplest approach of polling for handler to mark exit.
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			s := handle.TaskStatus()
			if s.State == drivers.TaskStateExited {
				ch <- handle.exitResult
			}
		}
	}
}

func (d *Driver) StopTask(taskID string, timeout time.Duration, signal string) error {
	handle, ok := d.tasks.Get(taskID)
	if !ok {
		return drivers.ErrTaskNotFound
	}

	if err := handle.shutdown(timeout); err != nil {
		return fmt.Errorf("executor Shutdown failed: %v", err)
	}

	return nil
}

func (d *Driver) DestroyTask(taskID string, force bool) error {
	handle, ok := d.tasks.Get(taskID)
	if !ok {
		return drivers.ErrTaskNotFound
	}

	if handle.IsRunning() && !force {
		return fmt.Errorf("cannot destroy running task")
	}

	if handle.IsRunning() {
		// grace period is chosen arbitrary here
		if err := handle.shutdown(1 * time.Minute); err != nil {
			handle.logger.Error("failed to destroy executor", "err", err)
		}
	}

	d.tasks.Delete(taskID)
	return nil
}

func (d *Driver) InspectTask(taskID string) (*drivers.TaskStatus, error) {
	handle, ok := d.tasks.Get(taskID)

	if !ok {
		return nil, drivers.ErrTaskNotFound
	}

	return handle.TaskStatus(), nil
}

func (d *Driver) TaskStats(ctx context.Context, taskID string, interval time.Duration) (<-chan *drivers.TaskResourceUsage, error) {
	handle, ok := d.tasks.Get(taskID)
	if !ok {
		return nil, drivers.ErrTaskNotFound
	}

	return handle.stats(ctx, interval)
}

func (d *Driver) TaskEvents(ctx context.Context) (<-chan *drivers.TaskEvent, error) {
	return d.eventer.TaskEvents(ctx)
}

func (d *Driver) SignalTask(taskID string, signal string) error {
	return fmt.Errorf("Zone driver does not support signals")
}

func (d *Driver) ExecTask(taskID string, cmd []string, timeout time.Duration) (*drivers.ExecTaskResult, error) {
	return nil, fmt.Errorf("Zone driver does not support exec")
}

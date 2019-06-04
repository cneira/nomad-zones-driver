package zone

import (
	"fmt"
	"git.wegmueller.it/illumos/go-zone/config"
	"github.com/hashicorp/nomad/plugins/drivers"
	"strings"
)

const (
	IpTypeShared    = 1
	IpTypeExclusive = 2
)

func (d *Driver) initializeContainer(cfg *drivers.TaskConfig, taskConfig TaskConfig) *config.Zone {
	containerName := fmt.Sprintf("%s-%s", cfg.Name, cfg.AllocID)
	z := config.New(containerName)
	z.Brand = taskConfig.Brand
	z.Zonepath = fmt.Sprintf("%s/%s", taskConfig.Zonepath, containerName)

	z.SetCPUShares(taskConfig.CpuShares)
	z.CappedMemory = config.NewMemoryCap(taskConfig.CappedMemory)
	z.SetMaxLockedMemory(taskConfig.LockedMemory)

	if len(taskConfig.SwapMemory) != 0 {
		z.SetMaxSwap(taskConfig.SwapMemory)
	}
	if len(taskConfig.ShmMemory) != 0 {
		z.SetMaxShmMemory(taskConfig.ShmMemory)
	}
	if len(taskConfig.ShmIds) != 0 {
		z.SetMaxShmIds(taskConfig.ShmIds)
	}
	if len(taskConfig.SemIds) != 0 {
		z.SetMaxSemIds(taskConfig.SemIds)
	}
	if len(taskConfig.MsgIds) != 0 {
		z.SetMaxMsgIds(taskConfig.MsgIds)
	}
	z.SetMaxLwps(taskConfig.Lwps)
	z.DedicatedCpu, _ = config.NewDedicatedCPU(taskConfig.DedicatedCpu)

	z.Networks = taskConfig.Networks
	z.IpType = IpTypeExclusive
	if strings.ToUpper(taskConfig.IpType) == "SHARED" {
		z.IpType = IpTypeShared
	}

	z.Attributes = taskConfig.Attributes
	z.Devices = taskConfig.Devices
	z.FileSystems = taskConfig.FileSystems
	return z
}

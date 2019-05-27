package zone

import (
	"fmt"
	"strings"

	"git.wegmueller.it/illumos/go-zone/config"
	"github.com/hashicorp/nomad/plugins/drivers"
)


const (
        IpTypeShared    IpType = 1
        IpTypeExclusive        = 2
)

func (d *Driver) initializeContainer(cfg *drivers.TaskConfig, taskConfig TaskConfig) *config.Zone {

	containerName := fmt.Sprintf("%s-%s", cfg.Name, cfg.AllocID)
	z := config.New(containerName)
	z.Brand = taskConfig.Brand
	z.Zonepath = fmt.Sprintf("%s/%s", taskConfig.Zonepath, containerName)
	z.SetCPUShares(taskConfig.CpuShares)
	z.CappedMemory = config.NewMemoryCap(taskConfig.Memory)
	z.Networks = taskConfig.Networks
	z.Attributes = taskConfig.Attributes
	z.IpType = IpTypeExclusive
	z.SetMaxLwps(taskConfig.Lwps)
	return z
}

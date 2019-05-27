package zone

import (
	"fmt"
	"strings"

	"git.wegmueller.it/illumos/go-zone/config"
	"github.com/hashicorp/nomad/plugins/drivers"
)

func (d *Driver) initializeContainer(cfg *drivers.TaskConfig, taskConfig TaskConfig) *config.Zone {

	containerName := fmt.Sprintf("%s-%s", cfg.Name, cfg.AllocID)
	z := config.New(containerName)
	z.Brand = taskConfig.Brand
	z.Zonepath = fmt.Sprintf("%s/%s", taskConfig.Zonepath, containerName)
	z.SetCPUShares(taskConfig.CpuShares)
	z.IpType = convert_to_IpType(TaskConfig.IpType)
	z.CappedMemory = config.NewMemoryCap(taskConfig.Memory)
	z.Networks = taskConfig.Networks
	z.Attributes = taskConfig.Attributes
	return z
}

/* TODO: validate fields before conversion
 */

func convert_to_IpType(string iptype) IpType {
	switch strings.ToLower(iptype) {
	case "shared":
		return IpTypeShared
	default:
		return  IpTypeExclusive

	}
}

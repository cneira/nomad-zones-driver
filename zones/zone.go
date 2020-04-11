/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 *
 * Copyright (c) 2018, Carlos Neira cneirabustos@gmail.com
 */

package zone

import (
	"crypto/rand"
	"fmt"
	"git.wegmueller.it/illumos/go-zone/config"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/plugins/drivers"
	"strings"
)

const (
	IpTypeShared    = 1
	IpTypeExclusive = 2
)

func simple_uuid() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("error calling rand.Read")
	}
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return uuid, nil
}

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

	z.Devices = taskConfig.Devices
	z.FileSystems = taskConfig.FileSystems
	z.Attributes = taskConfig.Attributes

	if len(taskConfig.Envars) != 0 {
		envars := config.Attribute{Name: "ENVARS", Type: "string", Value: string(taskConfig.Envars)}
		z.Attributes = append(z.Attributes, envars)
		d.logger.Info("driver_initialize_container", "Adding Envars", hclog.Fmt("%v+", z.Attributes))
	}

	if len(taskConfig.Docker) != 0 {

		env := config.Attribute{Name: "DOCKER", Type: "string", Value: string(taskConfig.Docker)}
		z.Attributes = append(z.Attributes, env)
		s := strings.Split(taskConfig.Docker, " ")
		d.logger.Info("Pulling image", "driver_initialize_container", hclog.Fmt("%v+", s))
		uuid, _ := simple_uuid()
		if len(s) > 1 {
			library, tag := s[0], s[1]
			name := strings.Split(library, "/")
			var libtag string

			if len(name) > 1 {
				library = name[1]
				libtag = s[0]
			} else {
				libtag = "library/" + s[0]

			}

			path := "/tmp/" + library + "-" + tag + "-" + uuid

			err := dockerpull(libtag, tag, path)

			if err == nil {
				img := config.Attribute{Name: "img", Type: "string", Value: path + ".tar.gz"}
				z.Attributes = append(z.Attributes, img)
				d.logger.Info("driver_initialize_container", "docker_pull", hclog.Fmt("%v+", z.Attributes))
			} else {
				d.logger.Info("driver_initialize_container", "docker_pull failed", hclog.Fmt("%v+", err))
			}

			m, err := docker_getconfig(libtag, tag)

			if err == nil {
				if val, ok := m["cmd"]; ok {
					cmd := config.Attribute{Name: "CMD", Type: "string", Value: string(val)}
					z.Attributes = append(z.Attributes, cmd)
				}
				if val, ok := m["env"]; ok {
					env := config.Attribute{Name: "ENV", Type: "string", Value: string(val)}
					z.Attributes = append(z.Attributes, env)
				}

				if val, ok := m["entrypoint"]; ok {
					env := config.Attribute{Name: "ENTRYPOINT", Type: "string", Value: string(val)}
					z.Attributes = append(z.Attributes, env)
				}
			}
		} else {

			d.logger.Info("Docker Registry Error check Docker value", "driver_initialize_container", hclog.Fmt("%v+", string(taskConfig.Docker)))
		}
	}
	d.logger.Info("taskConfig.Attributes", "driver_initialize_container", hclog.Fmt("%v+", z.Attributes))

	return z
}

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 *
 * Copyright (c) 2018, Carlos Neira cneirabustos@gmail.com
 */

package zone

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
	"os/exec"

	"github.com/ztrue/tracerr"
	zconfig "git.wegmueller.it/illumos/go-zone/config"
	"git.wegmueller.it/illumos/go-zone/lifecycle"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/plugins/drivers"
)

const (
	// containerMonitorIntv is the interval at which the driver checks if the
	// container is still running

	containerMonitorIntv = 2 * time.Second
	zoneStateRunning     = 86
)

type taskHandle struct {
	container zconfig.Zone
	logger    hclog.Logger

	// stateLock syncs access to all fields below
	stateLock sync.RWMutex

	taskConfig  *drivers.TaskConfig
	State       drivers.TaskState
	startedAt   time.Time
	completedAt time.Time
	exitResult  *drivers.ExitResult
}

func (h *taskHandle) TaskStatus() *drivers.TaskStatus {
	h.stateLock.RLock()
	defer h.stateLock.RUnlock()

	return &drivers.TaskStatus{
		ID:          h.taskConfig.ID,
		Name:        h.taskConfig.Name,
		State:       h.State,
		StartedAt:   h.startedAt,
		CompletedAt: h.completedAt,
		ExitResult:  h.exitResult,
	}
}

func (h *taskHandle) IsRunning() bool {
	h.stateLock.RLock()
	defer h.stateLock.RUnlock()
	return h.State == drivers.TaskStateRunning
}

func (h *taskHandle) run() {
	h.stateLock.Lock()
	if h.exitResult == nil {
		h.exitResult = &drivers.ExitResult{}
	}
	h.stateLock.Unlock()

	containerName := fmt.Sprintf("%s-%s", h.taskConfig.Name, h.taskConfig.AllocID)
	z := zconfig.New(containerName)
	mgr, err := lifecycle.NewManager(z)
	if err != nil {
		return
	}

	for mgr.GetZoneState() == zoneStateRunning {
		time.Sleep(containerMonitorIntv)
	}
	h.stateLock.Lock()
	defer h.stateLock.Unlock()

	h.State = drivers.TaskStateExited
	h.exitResult.ExitCode = 0
	h.exitResult.Signal = 0
	h.completedAt = time.Now()

}

/*
 * TODO: add cpu + memory stats from container
 */
func (h *taskHandle) stats(ctx context.Context, interval time.Duration) (<-chan *drivers.TaskResourceUsage, error) {
	return nil, nil
}

func (h *taskHandle) handleStats(ctx context.Context, ch chan *drivers.TaskResourceUsage, interval time.Duration) {
	defer close(ch)

}

func keysToVal(line string) (string, uint64, error) {
	tokens := strings.Split(line, " ")
	if len(tokens) != 2 {
		return "", 0, fmt.Errorf("line isn't a k/v pair")
	}
	key := tokens[0]
	val, err := strconv.ParseUint(tokens[1], 10, 64)
	return key, val, err
}

// shutdown shuts down the container, with `timeout` grace period
// before shutdown a zone.
func (h *taskHandle) shutdown(timeout time.Duration) error {
	containerName := fmt.Sprintf("%s-%s", h.taskConfig.Name, h.taskConfig.AllocID)
	z := zconfig.New(containerName)
	z.Brand = h.container.Brand
	z.Zonepath = h.container.Zonepath
	mgr, err := lifecycle.NewManager(z)
	if err != nil {
		return err
	}

	time.Sleep(timeout)
	if z.Brand == "lx" {
		var cmd *exec.Cmd
		cmd = exec.Command("zoneadm", "-z", containerName, "halt")

		if out, err := cmd.CombinedOutput(); err != nil {
			return tracerr.Wrap(fmt.Errorf("failed to run zoneadm -z %s shutdown: %s", containerName, out))
		}

		return nil

	} else {
		err = mgr.Shutdown(nil)

		if err != nil {
			return err
		}
	}

	return nil
}

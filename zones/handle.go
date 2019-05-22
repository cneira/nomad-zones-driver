package zone

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/plugins/drivers"
	"git.wegmueller.it/illumos/go-zone/config"
	"git.wegmueller.it/illumos/go-zone/lifecycle"
	zconfig "git.wegmueller.it/illumos/go-zone/config"
)

type taskHandle struct {
	container *config.Zone
	logger    hclog.Logger


	// stateLock syncs access to all fields below
	stateLock sync.RWMutex

	taskConfig  *drivers.TaskConfig
	State   drivers.TaskState
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
	
	containerName := fmt.Sprintf("%s-%s", h.taskConfig.Name,h.taskConfig.AllocID)
	z := zconfig.New(containerName)

	mgr, err := lifecycle.NewManager(z)
	if err != nil {
		return 
	}
	
	for mgr.GetZoneState() == 86 {
		time.Sleep(2 * time.Second)
	}
	h.stateLock.Lock()
	defer h.stateLock.Unlock()

	h.State = drivers.TaskStateExited
	h.exitResult.ExitCode = 0
	h.exitResult.Signal = 0
	h.completedAt = time.Now()

}

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
// before killing the container with SIGKILL.
func (h *taskHandle) shutdown(timeout time.Duration) error {
	return nil
}

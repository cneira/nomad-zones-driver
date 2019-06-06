package zone

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"git.wegmueller.it/illumos/go-zone/config"
	"github.com/hashicorp/nomad/plugins/drivers"
	hclog "github.com/hashicorp/go-hclog"
	"io"
	"net/http"
	"os"
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
	uuid := fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return uuid, nil
}

func dockerpull(library string, tag string, path string) error {
	resp, err := http.Get("https://auth.docker.io/token?service=registry.docker.io&scope=repository:library/" + library + ":pull")
	if err != nil {
		return fmt.Errorf("failed to get token")
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	token := result["token"].(string)
	req, err2 := http.NewRequest("GET", "https://registry-1.docker.io/v2/library/"+library+"/manifests/"+tag, nil)
	req.Header.Add("Authorization", "Bearer "+string(token))
	client := &http.Client{}
	resp, err2 = client.Do(req)
	if err2 != nil {
		return fmt.Errorf("Failed retrieving blobs")
	}

	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&result)
	blobs := result["fsLayers"].([]interface{})

	for _, v := range blobs {
		m := v.(map[string]interface{})
		for _, o := range m {
			url := "https://registry-1.docker.io/v2/library/" + library + "/blobs/" + o.(string)
			req, err := http.NewRequest("GET", url, nil)
			req.Header.Add("Authorization", "Bearer "+string(token))
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("Failed retrieving image blob: " + o.(string))
			}
			defer resp.Body.Close()
			// Create the file
			out, err := os.Create(path)
			if err != nil {
				return fmt.Errorf("Failed creating temporary image")
			}
			defer out.Close()

			// Write the body to file
			_, err = io.Copy(out, resp.Body)
		}
	}
	return nil
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

	if len(taskConfig.Docker) != 0 {
		s := strings.Split(taskConfig.Docker, "/")
			d.logger.Info("Pulling image", "driver_initialize_container", hclog.Fmt("%v+", s))
		uuid, _ := simple_uuid()
		if len(s) > 1 {
			library, tag := s[0], s[1]
			path := "/tmp/" + library + "-" + tag + "-" + uuid + ".gz"
			err := dockerpull(library, tag, path)
			if err != nil {
				img := config.Attribute{Name: "img", Type: "string", Value: path}
				z.Attributes = append(taskConfig.Attributes, img)
			}
		}
	}

	z.Attributes = taskConfig.Attributes

	return z
}

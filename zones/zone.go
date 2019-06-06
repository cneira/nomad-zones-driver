package zone

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"git.wegmueller.it/illumos/go-zone/config"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/plugins/drivers"
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
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return uuid, nil
}

func docker_getconfig(library string, tag string) (map[string]string, error) {
	respo, erro := http.Get("https://auth.docker.io/token?service=registry.docker.io&scope=repository:" + library + ":pull&service=registry.docker.io")
	if erro != nil {
		fmt.Println("Failed getting token")
		return nil, fmt.Errorf("failed to get token")
	}
	defer respo.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(respo.Body).Decode(&result)
	token := result["token"].(string)

	//GET DIGEST
	req, err2 := http.NewRequest("GET", "https://registry-1.docker.io/v2/"+library+"/manifests/"+tag, nil)
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.Header.Add("Authorization", "Bearer "+string(token))
	client := &http.Client{}
	resp, err2 := client.Do(req)
	if err2 != nil {
		fmt.Println("Failed getting digest")
		return nil, fmt.Errorf("Failed retrieving blobs")
	}

	var resultdigest map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&resultdigest)
	if resultdigest["config"] == nil {
		fmt.Println("Failed retriving digest ", resp.Header["Www-Authenticate"])
		return nil, fmt.Errorf("Failed retrieving digest")
	}
	config := resultdigest["config"].(map[string]interface{})
	digest := config["digest"].(string)
	defer resp.Body.Close()

	// GET CONTAINER CONFIG
	digesturl := "https://registry-1.docker.io/v2/" + library + "/blobs/" + digest

	req2, err3 := http.NewRequest("GET", digesturl, nil)
	if err3 != nil {
		fmt.Println("Failed retriving container config")
		return nil, fmt.Errorf("Failed retrieving blobs")
	}
	req2.Header.Add("Authorization", "Bearer "+string(token))
	req2.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	client2 := &http.Client{}
	resp2, err4 := client2.Do(req2)
	if err4 != nil {
		return nil, fmt.Errorf("Failed retrieving image blob: ")
	}

	var result3 map[string]interface{}
	json.NewDecoder(resp2.Body).Decode(&result3)
	container_config := result3["container_config"].(map[string]interface{})
	cmds := container_config["Cmd"].([]interface{})

	var execute []string
	m := make(map[string]string)

	if container_config["Entrypoint"] != nil {
		entrypoint := container_config["Entrypoint"].([]interface{})
		fmt.Printf("%s\n", entrypoint)
		m["entrypoint"] = fmt.Sprintf("%s", entrypoint)

	}

	env := container_config["Env"].([]interface{})

	if container_config["Cmd"] != nil {
		for _, v := range cmds {
			if strings.Contains(v.(string), "CMD") {
				execute = append(execute, v.(string))
			}
		}
	}

	defer resp.Body.Close()
	cmdargs := strings.Join(execute, " ")
	fmt.Println(cmdargs)
	fmt.Println(env)
	m["env"] = fmt.Sprintf("%s", env)
	m["cmd"] = fmt.Sprintf("%s", cmdargs)
	return m, nil
}

func dockerpull(library string, tag string, path string) error {
	resp, err := http.Get("https://auth.docker.io/token?service=registry.docker.io&scope=repository:" + library + ":pull")
	if err != nil {
		return fmt.Errorf("failed to get token")
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	token := result["token"].(string)
	req, err2 := http.NewRequest("GET", "https://registry-1.docker.io/v2/"+library+"/manifests/"+tag, nil)
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
			url := "https://registry-1.docker.io/v2/" + library + "/blobs/" + o.(string)
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
	z.Attributes = taskConfig.Attributes

	if len(taskConfig.Docker) != 0 {
		s := strings.Split(taskConfig.Docker, " ")
		d.logger.Info("Pulling image", "driver_initialize_container", hclog.Fmt("%v+", s))
		uuid, _ := simple_uuid()
		if len(s) > 1 {
			library, tag := s[0], s[1]
			name := strings.Split(library, "/")

			if len(name) > 1 {
				library = name[1]
			}

			path := "/tmp/" + library + "-" + tag + "-" + uuid + ".gz"

			libtag := s[0]

			d.logger.Info("driver_initialize_container", hclog.Fmt("library = %v+", library))
			d.logger.Info("driver_initialize_container", hclog.Fmt("path = %v+", path))
			d.logger.Info("driver_initialize_container", hclog.Fmt("libtag = %v+", libtag))

			err := dockerpull(libtag, tag, path)

			if err == nil {
				img := config.Attribute{Name: "img", Type: "string", Value: path}
				z.Attributes = append(z.Attributes, img)
				d.logger.Info("taskConfig.Attributes", "driver_initialize_container", hclog.Fmt("%v+", z.Attributes))
			}

			m, err := docker_getconfig(libtag, tag)

			if err == nil {
				if val, ok := m["cmd"]; ok {
					cmd := config.Attribute{Name: "Cmd", Type: "string", Value: string(val)}
					z.Attributes = append(z.Attributes, cmd)
				}
				if val, ok := m["env"]; ok {
					env := config.Attribute{Name: "Env", Type: "string", Value: string(val)}
					z.Attributes = append(z.Attributes, env)
				}
			}
		}
	}
	d.logger.Info("taskConfig.Attributes", "driver_initialize_container", hclog.Fmt("%v+", z.Attributes))

	return z
}

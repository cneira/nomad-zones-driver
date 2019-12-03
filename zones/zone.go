/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 *
 * Copyright (c) 2018, Carlos Neira cneirabustos@gmail.com
 */

package zone

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"git.wegmueller.it/illumos/go-zone/config"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/plugins/drivers"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
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

	var execute []string
	m := make(map[string]string)
	var re = regexp.MustCompile(`([\[|\]])`)

	if container_config["Entrypoint"] != nil {
		entrypoint := container_config["Entrypoint"].([]interface{})
		entry := fmt.Sprintf("%s", entrypoint)
		m["entrypoint"] = re.ReplaceAllString(entry, "")
	}

	if container_config["Env"] != nil {
		envi := container_config["Env"].([]interface{})
		env := fmt.Sprintf("%s", envi)
		m["env"] = re.ReplaceAllString(env, "")
	}

	if container_config["Cmd"] != nil {
		cmds := container_config["Cmd"].([]interface{})
		for _, v := range cmds {
			if strings.Contains(v.(string), "CMD") {
				scmd := strings.Replace(v.(string),`CMD`," ", -1)
				execute = append(execute, scmd)
			}
		}
		if len(execute) > 0 {
			cmdargs := strings.Join(execute, " ")
			m["cmd"] = fmt.Sprintf("%s", re.ReplaceAllString(cmdargs, ""))
		}
	}

	defer resp.Body.Close()

	return m, nil
}

func RemoveDuplicatesFromSlice(s []string) []string {
	m := make(map[string]bool)
	for _, item := range s {
		if _, ok := m[item]; ok {
		} else {
			m[item] = true
		}
	}

	var result []string
	for item, _ := range m {
		result = append(result, item)
	}
	return result
}

func dockerpull(library string, tag string, path string) error {
	resp, err := http.Get("https://auth.docker.io/token?service=registry.docker.io&scope=repository:" + library + ":pull")
	if err != nil {
		return fmt.Errorf("failed to get token from docker registry: %s", err)
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("could not decode response from registry: %s", err)	
	}
	token := result["token"].(string)

	req, manifestErr := http.NewRequest("GET", "https://registry-1.docker.io/v2/"+library+"/manifests/"+tag, nil)
	if err != nil {
		return fmt.Errorf("could not build request to get manifest: %s", err)	
	}
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.Header.Add("Authorization", "Bearer "+string(token))

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err2 = client.Do(req)
	if err2 != nil {
		return fmt.Errorf("Failed retrieving manifest: %s", err)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Failed to read manifest: %s", err)
	}

	defer resp.Body.Close()

	schema := make(map[string]interface{})
	json.Unmarshal(bodyBytes, &schema)
	blobs := schema["layers"].([]interface{})
	var gzblobs []string

	for _, v := range blobs {
		m := v.(map[string]interface{})
		for k, o := range m {
			if k == "digest" {
				gzblobs = append(gzblobs, o.(string))
			}
		}
	}

	//This should not be necessary acording to the OCI standards 
	dedupblobs := RemoveDuplicatesFromSlice(gzblobs)

	for _, blob := range dedupblobs {
		url := "https://registry-1.docker.io/v2/" + library + "/blobs/" + blob
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return fmt.Errorf("could not make request to get blob:%s err=%s", blob, err)	
		}
		req.Header.Add("Authorization", "Bearer "+string(token))
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("Failed retrieving image blob:%s err=%s ", blob, err)
		}

		out, err := os.Create("/tmp/" + blob + ".gz")
		if err != nil {
			return fmt.Errorf("Failed creating temporary image: %s")
		}

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return fmt.Errorf("could read body for blob:%s err=%s", blob, err)	
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			os.Mkdir(path, os.ModePerm)
		}
		args := []string{"xvfz", "/tmp/" + blob + ".gz", "-C", path}

		if err := exec.Command("gtar", args...).Run(); err != nil {
			return fmt.Errorf("error running gtar:%s", err, args)
		}

		cleanerr := os.Remove("/tmp/" + blob + ".gz")
		if cleanerr != nil {
			return fmt.Errorf("Failed cleaning up", cleanerr)
		}

		//TODO move to seperate function or closure so defer works resource leak in error case
		out.Close()
		resp.Body.Close()
	}
	cargs := []string{"cvfz", path + ".tar.gz", "-C", path, "."}
	if err := exec.Command("gtar", cargs...).Run(); err != nil {
		return fmt.Errorf("error running compress: %s:%s", err, cargs)
	}

	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("Failed cleaning up image dir %s", cleanerr)
	}

	return nil
}

func (d *Driver) initializeContainer(cfg *drivers.TaskConfig, taskConfig TaskConfig) (*config.Zone, error) {
	var containerName string
	if len(taskConfig.Zonename) != 0 {
		containerName = fmt.Sprintf("%s", taskConfig.Zonename)
	} else {
		containerName = fmt.Sprintf("%s-%s", cfg.Name, cfg.AllocID)
	}
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

	if len(taskConfig.Docker) != 0  {
		s := strings.Split(taskConfig.Docker, " ")
		d.logger.Info("Pulling image", "driver_initialize_container", hclog.Fmt("%v+", s))
		uuid, _ := simple_uuid()
		if len(s) > 1 {
			library, tag := s[0], s[1]
			name := strings.Split(library, "/")
			var libtag string
			if len(name) > 1 {
				library = name[1]
				libtag  = s[0]
			} else {
				libtag  = "library/" + s[0]
				library = name[0]
			}
			path := "/tmp/" + library + "-" + tag + "-" + uuid
			err := dockerpull(libtag, tag, path)
			if err == nil {
				img := config.Attribute{Name: "img", Type: "string", Value: path + ".tar.gz"}
				z.Attributes = append(z.Attributes, img)
				d.logger.Info("driver_initialize_container", "docker_pull", hclog.Fmt("%v+", z.Attributes))
			} else {
				d.logger.Info("driver_initialize_container", "docker_pull failed", hclog.Fmt("%v+", err))
				return nil, fmt.Errorf("Pulling image from docker.io failed: %s with error=%v+", taskConfig.Docker, err)
			}

			m, err := docker_getconfig(libtag, tag)

			if err == nil {
				if val, ok := m["cmd"]; ok {
					cmd := config.Attribute{Name: "cmd", Type: "string", Value: string(val)}
					z.Attributes = append(z.Attributes, cmd)
				}
				if val, ok := m["env"]; ok {
					if len(taskConfig.Envars) != 0 {
						val = string(val) + " " + taskConfig.Envars
					}  
					env := config.Attribute{Name: "env", Type: "string", Value: string(val)}
					z.Attributes = append(z.Attributes, env)
				}

				if val, ok := m["entrypoint"]; ok {
					env := config.Attribute{Name: "entrypoint", Type: "string", Value: string(val)}
					z.Attributes = append(z.Attributes, env)
				}
			} else {
				return nil, fmt.Errorf("Could not get entrypoint from docker image=%s %s err=%v+", libtag, tag, err)
			}
		}
	}
	d.logger.Info("taskConfig.Attributes", "driver_initialize_container", hclog.Fmt("%v+", z.Attributes))
	return z, nil
}

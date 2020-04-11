/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 *
 * Copyright (c) 2018, Carlos Neira cneirabustos@gmail.com
 */

package zone

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
)



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
				execute = append(execute, v.(string))
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
		return fmt.Errorf("failed to get token from docker registry")
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	token := result["token"].(string)

	req, err2 := http.NewRequest("GET", "https://registry-1.docker.io/v2/"+library+"/manifests/"+tag, nil)
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.Header.Add("Authorization", "Bearer "+string(token))

	client := &http.Client{}
	resp, err2 = client.Do(req)
	if err2 != nil {
		return fmt.Errorf("Failed retrieving blobs")
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Failed retrieving blobs")
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

	dedupblobs := RemoveDuplicatesFromSlice(gzblobs)

	for _, blob := range dedupblobs {
		url := "https://registry-1.docker.io/v2/" + library + "/blobs/" + blob
		req, err := http.NewRequest("GET", url, nil)
		req.Header.Add("Authorization", "Bearer "+string(token))
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("Failed retrieving image blob:%s err=%s ", blob, err)
		}

		out, err := os.Create("/tmp/" + blob + ".gz")
		if err != nil {
			return fmt.Errorf("Failed creating temporary image: %s")
		}

		_, err = io.Copy(out, resp.Body)

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

		out.Close()
		resp.Body.Close()
	}
	cargs := []string{"cvfz", path + ".tar.gz", "-C", path, "."}
	if err := exec.Command("gtar", cargs...).Run(); err != nil {
		return fmt.Errorf("error running compress: %s:%s", err, cargs)
	}

	cleanerr := os.RemoveAll(path)

	if cleanerr != nil {
		return fmt.Errorf("Failed cleaning up image dir %s", cleanerr)
	}

	return nil
}



package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	//"runtime"
	"net/http"
	"strings"
	"time"
)

type PackageInfo struct {
	PackageName    string `json:"package_name"`
	PackageVersion string `json:"pachage_version"`
	PackageRepo    string `json:"package_repo"`
}

type Result struct {
	Post struct {
		HostName string        `json:"host_name"`
		HostOs   string        `json:"host_os"`
		Packages []PackageInfo `json:"packages"`
	} `json:"post"`
}

func getOsName() string {
	cmd := exec.Command("cat", "/etc/issue")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cmd.Start()
	scanner := bufio.NewScanner(stdout)

	var os_name string
	i := 0
	for scanner.Scan() {
		//Get first line
		if i == 0 {
			os_name = scanner.Text()
			if err := scanner.Err(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			break
		}
	}
	return os_name
}

func getHostName() string {
	host_name, err := os.Hostname()
	if err != nil {
		fmt.Println("Get Hostname error:", err)
		os.Exit(1)
	}
	return host_name
}

func getInstalledList() []PackageInfo {
	var installed_list []PackageInfo

	cmd := exec.Command("yum", "list", "installed")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cmd.Start()

	lists := make([]string, 0, 500)
	scanner := bufio.NewScanner(stdout)
	i := 0
	for scanner.Scan() {
		//Exclude 2 line from stdout
		if i > 1 {
			lists = append(lists, scanner.Text())
			if err := scanner.Err(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		i = i + 1
	}

	list := strings.Join(lists, " ")
	rep := regexp.MustCompile(`[ ]+`)
	list = rep.ReplaceAllString(list, " ")
	list_slice := strings.Split(list, " ")

	var package_name string
	var package_version string
	var package_repo string

	n := 0
	for i := 0; i < len(list_slice); i++ {
		if n == 0 {
			package_name = list_slice[i]
			n = n + 1
		} else if n == 1 {
			package_version = list_slice[i]
			n = n + 1
		} else {
			package_repo = list_slice[i]
			n = 0
			installed_list = append(installed_list, PackageInfo{PackageName: package_name, PackageVersion: package_version, PackageRepo: package_repo})
		}
	}
	return installed_list
}

func httpPost(url string, param []byte) int {
	req, err := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer(param),
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: time.Duration(15 * time.Second)}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	status := resp.StatusCode
	defer resp.Body.Close()

	return status
}

func main() {
	var r Result

	r.Post.HostName = getHostName()
	r.Post.HostOs = getOsName()
	r.Post.Packages = getInstalledList()

	result_json, err := json.Marshal(r)
	if err != nil {
		fmt.Println("JSON Marshal error:", err)
		return
	}

	out := new(bytes.Buffer)
	json.Indent(out, result_json, "", "    ")
	fmt.Println(out.String())

	resp := httpPost("http://10.0.2.2:4000/api/posts", result_json)
	fmt.Println(resp)
}

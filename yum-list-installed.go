package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	//"fmt"
	"github.com/comail/colog"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type YumInfo struct {
	PackageName    string `json:"package_name"`
	PackageVersion string `json:"pachage_version"`
	PackageRepo    string `json:"package_repo"`
}

type Result struct {
	Post struct {
		HostName string    `json:"host_name"`
		HostOs   string    `json:"host_os"`
		Packages []YumInfo `json:"packages"`
	} `json:"post"`
}

func getOsName() string {
	cmd := exec.Command("cat", "/etc/issue")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalln("error: cann't get os name", err)
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
				log.Fatalln("error: cann't get os name from scanner", err)
			}
			break
		}
	}
	log.Println("info: get os name", os_name)
	return os_name
}

func getHostName() string {
	host_name, err := os.Hostname()
	if err != nil {
		log.Fatalln("error: cann't get host name", err)
	}
	log.Println("info: get host name", host_name)
	return host_name
}

func getInstalledList() []YumInfo {
	var installed_list []YumInfo

	cmd := exec.Command("yum", "list", "installed")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalln("error: cann't execute yum command", err)
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
				log.Fatalln("error: cann't get scanner from yum command", err)
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
			installed_list = append(installed_list, YumInfo{PackageName: package_name, PackageVersion: package_version, PackageRepo: package_repo})
		}
	}
	log.Println("info: success to get yum list installed")
	return installed_list
}

func httpPost(url string, param []byte) int {
	req, err := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer(param),
	)
	if err != nil {
		log.Fatalln("error: cann't create http request", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: time.Duration(5 * time.Second)}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln("error: cann't execute http request", err)
	}
	status := resp.StatusCode
	defer resp.Body.Close()

	if status >= 400 {
		log.Println("warn: fatal to http post request", status)
	} else {
		log.Println("info: success to http post request", status)
	}
	return status
}

func main() {
	var r Result
	colog.Register()

	r.Post.HostName = getHostName()
	r.Post.HostOs = getOsName()
	r.Post.Packages = getInstalledList()

	result_json, err := json.Marshal(r)
	if err != nil {
		log.Println("error: JSON Marshal error", err)
	}

	//out := new(bytes.Buffer)
	//json.Indent(out, result_json, "", "    ")
	//fmt.Println(out.String())

	httpPost("http://10.0.2.2:4000/api/posts", result_json)
}
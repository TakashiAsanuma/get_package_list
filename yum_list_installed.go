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
	"strings"
)

type PackageInfo struct {
	PackageName    string `json:"package_name"`
	PackageVersion string `json:"pachage_version"`
	PackageRepo    string `json:"package_repo"`
}

type Resultslice struct {
	HostName string
	HostOs   string
	Packages []PackageInfo
}

func get_os_name() string {
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

func get_host_name() string {
	host_name, err := os.Hostname()
	if err != nil {
		fmt.Println("Get Hostname error:", err)
		os.Exit(1)
	}

	return host_name
}

func main() {
	var r Resultslice

	r.HostName = get_host_name()
	r.HostOs = get_os_name()

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
			r.Packages = append(r.Packages, PackageInfo{PackageName: package_name, PackageVersion: package_version, PackageRepo: package_repo})
		}
	}

	jsonBytes, err := json.Marshal(r)
	if err != nil {
		fmt.Println("JSON Marshal error:", err)
		return
	}

	out := new(bytes.Buffer)
	json.Indent(out, jsonBytes, "", "    ")
	fmt.Println(out.String())
}

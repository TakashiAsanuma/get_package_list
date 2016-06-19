package main
import (
    "bufio"
    "fmt"
    "os"
    "os/exec"
    "encoding/json"
    "bytes"
    "runtime"
    "strings"
    "regexp"
)

type Stats struct {
        HostName       string `json:"host_name"`
        HostOs         string `json:"host_os"`
        PackageList    string `json:"package_list"`
}

type PackageInfo struct {
        PackageName    string `json:"package_name"`
        PackageVersion string `json:"pachage_version"`
        PackageRepo    string `json:"package_repo"`
}

func main() {
    cmd := exec.Command("yum", "list", "installed")
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    hostname, err := os.Hostname()
    if err != nil {
        fmt.Println("Get Hostname error:", err)
        return
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
        i = i+1
    }

    list:= strings.Join(lists, " ")
    rep := regexp.MustCompile(`[ ]+`)
    list = rep.ReplaceAllString(list, " ")

    list_slice := strings.Split(list, " ")

    var package_name string
    var package_version string
    var package_repo string

    n :=0
    for i := 0; i < len(list_slice); i++ {
        if n == 0 {
            package_name = list_slice[i]
            n = n+1
        } else if n == 1 {
            package_version = list_slice[i]
            n = n+1
        } else {
            package_repo = list_slice[i]
            n = 0
            package_info := &PackageInfo{
                    PackageName:       package_name,
                    PackageVersion:    package_version,
                    PackageRepo:       package_repo,
            }
            fmt.Println(package_info)
        }
    }

    stats := &Stats{
            HostName:       hostname,
            HostOs:         runtime.GOOS,
            PackageList:    list,
    }

    jsonBytes, err := json.Marshal(stats)
    if err != nil {
        fmt.Println("JSON Marshal error:", err)
        return
    }

    out := new(bytes.Buffer)
    json.Indent(out, jsonBytes, "", "    ")
    fmt.Println(out.String())
}

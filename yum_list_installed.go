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

    lists := make([]string, 0, 10)
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

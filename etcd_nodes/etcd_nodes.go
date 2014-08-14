package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/coreos/go-etcd/etcd"
)

type KeyValue map[string]interface{}

type KeyValueGroup []KeyValue // a slice of KeyValue's

func checkForErrors(err error) {
	if err != nil {
		panic(err)
		os.Exit(2)
	}
}

func getDockerHostIP() (ip string) {
	cmd := fmt.Sprintf("netstat -nr | grep '^0\\.0\\.0\\.0' | awk '{print $2}'")
	out, err := exec.Command("sh", "-c", cmd).Output()
	checkForErrors(err)

	return string(out)
}

func getDiscoveryHost(host string, port string) (discoveryHost string) {
	return fmt.Sprintf("http://%s:%s", host, port)
}

func main() {
	// Local etcd API host & port
	port := "7001"
	discoveryHost := getDiscoveryHost(getDockerHostIP(), port)

	// Request path listing etcd machines in cluster
	discoveryPath := "admin/machines"

	// Connect & send request to local etcd API to retrieve machines JSON
	client := etcd.NewClient([]string{discoveryHost})
	req := etcd.NewRawRequest("GET", discoveryPath, nil, nil)
	rawRespPtr, err := client.SendRequest(req)
	checkForErrors(err)
	jsonResponse := rawRespPtr.Body

	var machines KeyValueGroup
	err = json.Unmarshal(jsonResponse, &machines)
	checkForErrors(err)

	for _, machine := range machines {
		log.Printf("%s -- %s -- %s -- %s\n", machine["name"], machine["state"], machine["clientURL"], machine["peerURL"])
	}
}

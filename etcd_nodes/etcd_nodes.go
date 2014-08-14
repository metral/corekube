package main

import (
	"encoding/json"
	"log"

	"github.com/coreos/go-etcd/etcd"
)

type ETCDMachine map[string]interface{}

type ETCDMachines []ETCDMachine

func checkForErrors(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	discoveryHost := "http://127.0.0.1:7001"
	discoveryPath := "admin/machines"

	// Connect to the etcd discovery to pull the nodes
	client := etcd.NewClient([]string{discoveryHost})
	req := etcd.NewRawRequest("GET", discoveryPath, nil, nil)
	rawRespPtr, err := client.SendRequest(req)
	checkForErrors(err)

	jsonResponse := rawRespPtr.Body

	var etcdMachines ETCDMachines
	err = json.Unmarshal(jsonResponse, &etcdMachines)

	checkForErrors(err)

	for _, etcdMachine := range etcdMachines {
		log.Printf("%s -- %s -- %s -- %s\n", etcdMachine["name"], etcdMachine["state"], etcdMachine["clientURL"], etcdMachine["peerURL"])
	}

}

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type KeyValue map[string]interface{}

type KeyValueGroup []KeyValue // a slice of KeyValue's

type EtcdMachine struct {
	name      string
	state     string
	clientURL string
	peerURL   string
}

type EtcdMachineGroup []EtcdMachine // a slice of EtcdMachine's

// Modify EtcdMachine
func (e *EtcdMachine) SetProperties(machineData KeyValue) {
	e.name = machineData["name"].(string)
	e.state = machineData["state"].(string)
	e.clientURL = machineData["clientURL"].(string)
	e.peerURL = machineData["peerURL"].(string)
}

// Retrieve string of EtcdMachine
func (e *EtcdMachine) String() string {
	output := fmt.Sprintf("Name: %s | State: %s | ClientURL: %s | PeerURL: %s",
		e.name,
		e.state,
		e.clientURL,
		e.peerURL,
	)
	return output
}

// Check for errors and panic, if found
func checkForErrors(err error) {
	if err != nil {
		log.Fatal("%s", err)
	}
}

// Get the IP address of the docker host as this is run from within container
func getDockerHostIP() string {
	cmd := fmt.Sprintf("netstat -nr | grep '^0\\.0\\.0\\.0' | awk '{print $2}'")
	out, err := exec.Command("sh", "-c", cmd).Output()
	checkForErrors(err)

	ip := string(out)
	ip = strings.Replace(ip, "\n", "", -1)
	return ip
}

// Compose the etcd API host:port location
func getEtcdAPI(host string, port string) string {
	return fmt.Sprintf("http://%s:%s", host, port)
}

func httpGetRequest(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal("%s", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal("%s", err)
	}

	return body
}

func machineRequestAndParse(url string, machineDataGroup *KeyValueGroup) {
	jsonResponse := httpGetRequest(url)

	// Decode the JSON returned
	err := json.Unmarshal(jsonResponse, &machineDataGroup)
	checkForErrors(err)
}

func waitForMachines(
	machineDataGroup *KeyValueGroup, expectedMachineCount int) {

	// Local etcd API host & port
	port := "7001"
	etcdAPI := getEtcdAPI(getDockerHostIP(), port)

	// Request path listing etcd machines in cluster
	etcdAPIPath := "v2/admin/machines"
	url := fmt.Sprintf("%s/%s", etcdAPI, etcdAPIPath)

	// Issue request to get machines & parse it. Sleep if cluster not ready yet
	machineRequestAndParse(url, machineDataGroup)
	for len(*machineDataGroup) < expectedMachineCount {
		machineRequestAndParse(url, machineDataGroup)
		log.Printf(
			"Waiting for all (%d) machines to join cluster...",
			expectedMachineCount)
		time.Sleep(1 * time.Second)
	}
}

func getMachines(machines *EtcdMachineGroup, expectedMachineCount int) {

	// Wait for all machines in the expected count to join cluster
	var machineDataGroup KeyValueGroup
	waitForMachines(&machineDataGroup, expectedMachineCount)

	// Use machine data to create local objects of the etcd machines
	for _, machineData := range machineDataGroup {
		machine := EtcdMachine{}
		machine.SetProperties(machineData)
		*machines = append(*machines, machine)
	}

}

func getState(machines *EtcdMachineGroup) string {
	hostname := os.Getenv("DOCKERHOST_HOSTNAME")
	//log.Printf("hostname env: %s", hostname)
	hostname = strings.Split(hostname, ".")[0]

	for _, machine := range *machines {
		if machine.name == hostname {
			return machine.state
		}
	}

	return ""
}

func Usage() {
	fmt.Printf("Usage: %s\n", os.Args[0])
	flag.PrintDefaults()
}

func setupFlags() int {
	expectedMachineCount :=
		flag.Int("machine_count", 0, "Expected number of machines in cluster")
	flag.Parse()

	return *expectedMachineCount
}

// Access the CoreOS / docker etcd API to extract machine information
func main() {
	expectedMachineCount := setupFlags()

	if expectedMachineCount <= 0 {
		Usage()
		os.Exit(2)
	}

	machines := EtcdMachineGroup{}
	getMachines(&machines, expectedMachineCount)

	// TODO: delete
	for _, machine := range machines {
		log.Printf("%s\n", machine.String())
	}

	state := getState(&machines)
	// TODO: delete
	log.Printf("%s", state)
}

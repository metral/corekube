package lib

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

// Check for errors and panic, if found
func checkForErrors(err error) {
	if err != nil {
		log.Fatal("Error:\n%s", err)
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

func getFullAPIURL(port, etcdAPIPath string) string {
	etcdAPI := getEtcdAPI(getDockerHostIP(), port)
	url := fmt.Sprintf("%s/%s", etcdAPI, etcdAPIPath)
	return url
}

func WaitForFleetMachines(
	fleetMachines *FleetMachines, expectedMachineCount int) {

	// Issue request to get machines & parse it. Sleep if cluster not ready yet
	url := getFullAPIURL("4001", "v2/keys/_coreos.com/fleet/machines")
	jsonResponse := httpGetRequest(url)
	err := json.Unmarshal(jsonResponse, fleetMachines)
	checkForErrors(err)
	totalMachines := len(fleetMachines.Node.Nodes)

	for totalMachines < expectedMachineCount {
		log.Printf("Waiting for all (%d) machines to be available "+
			"in fleet. Currently at: (%d)",
			expectedMachineCount, totalMachines)
		time.Sleep(1 * time.Second)

		jsonResponse := httpGetRequest(url)
		err := json.Unmarshal(jsonResponse, fleetMachines)
		checkForErrors(err)
		totalMachines = len(fleetMachines.Node.Nodes)
	}
}

func WaitForFleetMachineMetadata(
	fleetMachinesNodeNodesValue *FleetMachinesNodeNodesValue,
	fleetMachineObjectNodeValue *FleetMachineObjectNodeValue,
	expectedMachineCount int) {

	// Issue request to get machines & parse it. Sleep if cluster not ready yet
	id := strings.Split(fleetMachinesNodeNodesValue.Key, "fleet/machines/")[1]
	path := fmt.Sprintf("v2/keys/_coreos.com/fleet/machines/%s/object", id)

	url := getFullAPIURL("4001", path)
	jsonResponse := httpGetRequest(url)

	var fleetMachineObject FleetMachineObject
	err := json.Unmarshal(jsonResponse, &fleetMachineObject)
	checkForErrors(err)

	err = json.Unmarshal(
		[]byte(fleetMachineObject.Node.Value), &fleetMachineObjectNodeValue)
	checkForErrors(err)

	for len(fleetMachineObjectNodeValue.Metadata) == 0 ||
		fleetMachineObjectNodeValue.Metadata["kubernetes_role"] == nil {
		log.Printf("Waiting for machine (%s) metadata to be available "+
			"in fleet...", fleetMachineObjectNodeValue.ID)
		time.Sleep(1 * time.Second)

		err = json.Unmarshal(
			[]byte(fleetMachineObject.Node.Value), &fleetMachineObjectNodeValue)
		checkForErrors(err)

	}
}

func Usage() {
	fmt.Printf("Usage: %s\n", os.Args[0])
	flag.PrintDefaults()
}

func SetupFlags() int {
	expectedMachineCount :=
		flag.Int("machine_count", 0, "Expected number of machines in cluster")
	flag.Parse()

	return *expectedMachineCount
}

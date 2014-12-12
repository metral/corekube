package lib

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var ETCD_API_VERSION string = "v2"
var ETCD_CLIENT_PORT string = "4001"

type MinionsResult struct {
	Kind              string  `json:"kind,omitempty"`
	CreationTimestamp string  `json:"creationTimestamp,omitempty"`
	SelfLink          string  `json:"selfLink,omitempty"`
	APIVersion        string  `json:"apiVersion,omitempty"`
	Minions           Minions `json:"minions,omitempty"`
}

type Minions []Minion
type Minion struct {
	ID                string `json:"id,omitempty"`
	UID               string `json:"uid,omitempty"`
	CreationTimestamp string `json:"creationTimestamp,omitempty"`
	SelfLink          string `json:"selfLink,omitempty"`
	ResourceVersion   int    `json:"resourceVersion,omitempty"`
	HostIP            string `json:"hostIP,omitempty"`
	resources         map[interface{}]interface{}
}

// Check for errors and panic, if found
func checkForErrors(err error) {
	if err != nil {
		pc, fn, line, _ := runtime.Caller(1)
		msg := fmt.Sprintf("[Error] in %s[%s:%d] %v",
			runtime.FuncForPC(pc).Name(), fn, line, err)
		log.Fatal(msg)
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
	checkForErrors(err)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	checkForErrors(err)

	return body
}

func httpPutRequest(urlStr string, data []byte) *http.Response {
	var req *http.Request

	req, _ = http.NewRequest("PUT", urlStr, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	checkForErrors(err)

	defer resp.Body.Close()

	return resp
}

func httpPutRequestRedirect(urlStr string, data string) {
	var req *http.Request
	req, _ = http.NewRequest("PUT", urlStr, bytes.NewBufferString(data))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data)))

	client := &http.Client{}
	resp, err := client.Do(req)
	checkForErrors(err)

	// if resp is TemporaryRedirect, set the new leader and retry
	if resp.StatusCode == http.StatusTemporaryRedirect {
		u, err := resp.Location()

		if err != nil {
			checkForErrors(err)
		} else {
			//client.cluster.updateLeaderFromURL(u)
			httpPutRequestRedirect(u.String(), data)
		}
		resp.Body.Close()
	}
}

func getFullAPIURL(port, etcdAPIPath string) string {
	etcdAPI := getEtcdAPI(getDockerHostIP(), port)
	url := fmt.Sprintf("%s/%s", etcdAPI, etcdAPIPath)
	return url
}

func getFleetMachines(fleetResult *Result) {
	path := fmt.Sprintf("%s/keys/_coreos.com/fleet/machines", ETCD_API_VERSION)
	url := getFullAPIURL(ETCD_CLIENT_PORT, path)
	jsonResponse := httpGetRequest(url)
	err := json.Unmarshal(jsonResponse, fleetResult)
	checkForErrors(err)

	removeOverlord(&fleetResult.Node.Nodes)
}

func getMachinesDeployed() []string {
	var machinesDeployedResult NodeResult

	path := fmt.Sprintf("%s/keys/deployed", ETCD_API_VERSION)
	urlStr := getFullAPIURL(ETCD_CLIENT_PORT, path)

	jsonResponse := httpGetRequest(urlStr)
	err := json.Unmarshal(jsonResponse, &machinesDeployedResult)
	checkForErrors(err)

	var machinesDeployed []string
	var machinesDeployedBytes []byte = []byte(machinesDeployedResult.Node.Value)
	err = json.Unmarshal(machinesDeployedBytes, &machinesDeployed)
	checkForErrors(err)

	return machinesDeployed
}

func machineDeployed(id string) bool {
	deployed := false
	machineIDs := getMachinesDeployed()

	for _, machineID := range machineIDs {
		if machineID == id {
			deployed = true
		}
	}

	return deployed
}

func setMachineDeployed(id string) {
	path := fmt.Sprintf("%s/keys/deployed/", ETCD_API_VERSION)
	urlStr := getFullAPIURL(ETCD_CLIENT_PORT, path)
	data := ""

	switch id {
	case "":
		emptySlice := []string{}
		dataJSON, _ := json.Marshal(emptySlice)
		data = fmt.Sprintf("value=%s", dataJSON)
	default:
		machinesDeployed := getMachinesDeployed()
		machinesDeployed = append(machinesDeployed, id)
		dataJSON, _ := json.Marshal(machinesDeployed)
		data = fmt.Sprintf("value=%s", dataJSON)
	}

	httpPutRequestRedirect(urlStr, data)
}
func isK8sMachine(
	fleetMachine, master *FleetMachine,
	minions *FleetMachines) bool {

	role := fleetMachine.Metadata["kubernetes_role"]

	switch role {
	case "master":
		*master = *fleetMachine
		return true
	case "minion":
		*minions = append(*minions, *fleetMachine)
		return true
	}
	return false
}

func removeOverlord(nodes *ResultNodes) {
	var fleetMachine FleetMachine
	n := *nodes

	for i, node := range n {
		WaitForMetadata(&node, &fleetMachine)
		if fleetMachine.Metadata["kubernetes_role"] == "overlord" {
			n = append(n[:i], n[i+1:]...)
			*nodes = n
			break
		}
	}

}

func registerMinions(master *FleetMachine, minions *FleetMachines) {

	// Get registered minions, if any
	endpoint := fmt.Sprintf("http://%s:%s", master.PublicIP, K8S_API_PORT)
	masterAPIurl := fmt.Sprintf("%s/api/%s/minions", endpoint, K8S_API_VERSION)
	jsonResponse := httpGetRequest(masterAPIurl)
	m := *minions

	var minionsResult MinionsResult
	err := json.Unmarshal(jsonResponse, &minionsResult)
	checkForErrors(err)

	// See if minions discovered have been registered. If not, register
	for _, minion := range m {
		registered := false
		for _, registeredMinion := range minionsResult.Minions {
			if registeredMinion.HostIP == minion.PublicIP {
				registered = true
			}
		}

		if !registered {
			register(endpoint, minion.PublicIP)
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func Run(fleetResult *Result) {
	master := FleetMachine{}
	minions := FleetMachines{}

	setMachineDeployed("")
	getFleetMachines(fleetResult)
	totalMachines := len(fleetResult.Node.Nodes)

	// Get Fleet machines
	for {
		log.Printf("------------------------------------------------")
		log.Printf("Current # of machines discovered: (%d)\n", totalMachines)

		// Get Fleet machines metadata
		var fleetMachine FleetMachine
		for _, resultNode := range fleetResult.Node.Nodes {
			WaitForMetadata(&resultNode, &fleetMachine)

			if !machineDeployed(fleetMachine.ID) {
				log.Printf("------------------------------------------------")
				log.Printf("Found machine:\n")
				fleetMachine.PrintString()

				if isK8sMachine(&fleetMachine, &master, &minions) {
					createdFiles := createUnitFiles(&fleetMachine)
					for _, file := range createdFiles {
						if !unitFileCompleted(file) {
							startUnitFile(file)
							waitUnitFileComplete(file)
						}
					}
				}
				setMachineDeployed(fleetMachine.ID)
			}
		}

		registerMinions(&master, &minions)

		time.Sleep(1 * time.Second)
		getFleetMachines(fleetResult)
		totalMachines = len(fleetResult.Node.Nodes)
	}
}

func WaitForMetadata(
	resultNode *ResultNode,
	fleetMachine *FleetMachine,
) {

	// Issue request to get machines & parse it. Sleep if cluster not ready yet
	id := strings.Split(resultNode.Key, "fleet/machines/")[1]
	path := fmt.Sprintf(
		"%s/keys/_coreos.com/fleet/machines/%s/object", ETCD_API_VERSION, id)

	url := getFullAPIURL(ETCD_CLIENT_PORT, path)
	jsonResponse := httpGetRequest(url)

	var nodeResult NodeResult
	err := json.Unmarshal(jsonResponse, &nodeResult)
	checkForErrors(err)

	err = json.Unmarshal(
		[]byte(nodeResult.Node.Value), &fleetMachine)
	checkForErrors(err)

	for len(fleetMachine.Metadata) == 0 ||
		fleetMachine.Metadata["kubernetes_role"] == nil {
		log.Printf("Waiting for machine (%s) metadata to be available "+
			"in fleet...", fleetMachine.ID)
		time.Sleep(1 * time.Second)

		err = json.Unmarshal(
			[]byte(nodeResult.Node.Value), &fleetMachine)
		checkForErrors(err)

	}
}

func FindInfoForRole(
	role string,
	fleetMachines *[]FleetMachine) []string {
	var machines []string

	for _, fleetMachine := range *fleetMachines {
		if fleetMachine.Metadata["kubernetes_role"] == role {
			machines = append(machines, fleetMachine.PublicIP)
		}
	}

	return machines
}

func Usage() {
	fmt.Printf("Usage: %s\n", os.Args[0])
	flag.PrintDefaults()
}

package lib

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/metral/goutils"
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

// Get the IP address of the docker host as this is run from within container
func getDockerHostIP() string {
	cmd := fmt.Sprintf("netstat -nr | grep '^0\\.0\\.0\\.0' | awk '{print $2}'")
	out, err := exec.Command("sh", "-c", cmd).Output()
	goutils.CheckForErrors(goutils.ErrorParams{Err: err, CallerNum: 1})

	ip := string(out)
	ip = strings.Replace(ip, "\n", "", -1)
	return ip
}

// Compose the etcd API host:port location
func getEtcdAPI(host string, port string) string {
	return fmt.Sprintf("http://%s:%s", host, port)
}

func getFullAPIURL(port, etcdAPIPath string) string {
	etcdAPI := getEtcdAPI(getDockerHostIP(), port)
	url := fmt.Sprintf("%s/%s", etcdAPI, etcdAPIPath)
	return url
}

func getFleetMachines(fleetResult *Result) {
	path := fmt.Sprintf("%s/keys/_coreos.com/fleet/machines", ETCD_API_VERSION)
	url := getFullAPIURL(ETCD_CLIENT_PORT, path)
	jsonResponse := goutils.HttpGetRequest(url)
	err := json.Unmarshal(jsonResponse, fleetResult)
	goutils.CheckForErrors(goutils.ErrorParams{Err: err, CallerNum: 1})

	removeOverlord(&fleetResult.Node.Nodes)
}

func getMachinesSeen() []string {
	var machinesSeenResult NodeResult

	path := fmt.Sprintf("%s/keys/seen", ETCD_API_VERSION)
	urlStr := getFullAPIURL(ETCD_CLIENT_PORT, path)

	jsonResponse := goutils.HttpGetRequest(urlStr)
	err := json.Unmarshal(jsonResponse, &machinesSeenResult)
	goutils.CheckForErrors(goutils.ErrorParams{Err: err, CallerNum: 1})

	var machinesSeen []string
	var machinesSeenBytes []byte = []byte(machinesSeenResult.Node.Value)
	err = json.Unmarshal(machinesSeenBytes, &machinesSeen)
	goutils.CheckForErrors(goutils.ErrorParams{Err: err, CallerNum: 1})

	return machinesSeen
}

func machineSeen(allMachinesSeen []string, id string) bool {
	seen := false

	for _, machineSeen := range allMachinesSeen {
		if machineSeen == id {
			seen = true
		}
	}

	return seen
}

func setMachinesSeen(machines []string) {
	path := fmt.Sprintf("%s/keys/seen", ETCD_API_VERSION)
	urlStr := getFullAPIURL(ETCD_CLIENT_PORT, path)
	data := ""

	switch machines {
	case nil:
		emptySlice := []string{}
		dataJSON, _ := json.Marshal(emptySlice)
		data = fmt.Sprintf("value=%s", dataJSON)
	default:
		dataJSON, _ := json.Marshal(machines)
		data = fmt.Sprintf("value=%s", dataJSON)
	}

	goutils.HttpPutRequestRedirect(urlStr, data)
}
func isK8sMachine(
	fleetMachine,
	master *FleetMachine,
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
	jsonResponse := goutils.HttpGetRequest(masterAPIurl)
	m := *minions

	var minionsResult MinionsResult
	err := json.Unmarshal(jsonResponse, &minionsResult)
	goutils.CheckForErrors(goutils.ErrorParams{Err: err, CallerNum: 1})

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

func Run() {
	fleetResult := Result{}
	var f *Result = &fleetResult
	master := FleetMachine{}
	minions := FleetMachines{}

	setMachinesSeen([]string{})
	time.Sleep(1 * time.Second)

	// Get Fleet machines
	for {
		getFleetMachines(f)
		allMachinesSeen := getMachinesSeen()
		totalMachines := len(f.Node.Nodes)
		log.Printf("------------------------------------------------")
		log.Printf("Current # of machines discovered: (%d)\n", totalMachines)

		// Get Fleet machines metadata
		var fleetMachine FleetMachine
		for _, resultNode := range f.Node.Nodes {
			WaitForMetadata(&resultNode, &fleetMachine)

			if !machineSeen(allMachinesSeen, fleetMachine.ID) {
				log.Printf("------------------------------------------------")
				log.Printf("Found machine:\n")
				fleetMachine.PrintString()

				if isK8sMachine(&fleetMachine, &master, &minions) {
					allMachinesSeen = append(allMachinesSeen, fleetMachine.ID)
				}
				createdFiles := createUnitFiles(&fleetMachine)
				for _, file := range createdFiles {
					if !unitFileCompleted(file) {
						startUnitFile(file)
						waitUnitFileComplete(file)
					}
				}
			}
		}

		setMachinesSeen(allMachinesSeen)
		registerMinions(&master, &minions)
		time.Sleep(1 * time.Second)
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
	jsonResponse := goutils.HttpGetRequest(url)

	var nodeResult NodeResult
	err := json.Unmarshal(jsonResponse, &nodeResult)
	goutils.CheckForErrors(goutils.ErrorParams{Err: err, CallerNum: 1})

	err = json.Unmarshal(
		[]byte(nodeResult.Node.Value), &fleetMachine)
	goutils.CheckForErrors(goutils.ErrorParams{Err: err, CallerNum: 1})

	for len(fleetMachine.Metadata) == 0 ||
		fleetMachine.Metadata["kubernetes_role"] == nil {
		log.Printf("Waiting for machine (%s) metadata to be available "+
			"in fleet...", fleetMachine.ID)
		time.Sleep(1 * time.Second)

		err = json.Unmarshal(
			[]byte(nodeResult.Node.Value), &fleetMachine)
		goutils.CheckForErrors(goutils.ErrorParams{Err: err, CallerNum: 1})

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

package lib

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/metral/goutils"
)

var ETCD_API_VERSION string = "v2"
var ETCD_CLIENT_PORT string = "4001"

type Result struct {
	Action string
	Node   ResultNode
}

type ResultNodes []ResultNode
type ResultNode struct {
	Key           string
	Dir           bool
	Nodes         ResultNodes
	ModifiedIndex int
	CreatedIndex  int
}

type NodeResult struct {
	Action string
	Node   Node
}

type Node struct {
	Key           string
	Value         string
	Expiration    string
	Ttl           int
	ModifiedIndex int
	CreatedIndex  int
}

// Compose the etcd API host:port location
func getEtcdAPI(host string, port string) string {
	return fmt.Sprintf("http://%s:%s", host, port)
}

func getFleetMachines(fleetResult *Result) {
	path := fmt.Sprintf("%s/keys/_coreos.com/fleet/machines", ETCD_API_VERSION)
	url := getFullAPIURL(ETCD_CLIENT_PORT, path)

	headers := map[string]string{
		"Content-Type": "application/json",
	}
	p := goutils.HttpRequestParams{
		HttpRequestType: "GET",
		Url:             url,
		Headers:         headers,
	}

	_, jsonResponse := goutils.HttpCreateRequest(p)
	err := json.Unmarshal(jsonResponse, fleetResult)
	goutils.CheckForErrors(goutils.ErrorParams{Err: err, CallerNum: 1})

	removeOverlord(&fleetResult.Node.Nodes)
}

func getFullAPIURL(port, etcdAPIPath string) string {
	etcdAPI := getEtcdAPI(getDockerHostIP(), port)
	url := fmt.Sprintf("%s/%s", etcdAPI, etcdAPIPath)
	return url
}

func getMachinesSeen() []string {
	var machinesSeenResult NodeResult

	path := fmt.Sprintf("%s/keys/seen", ETCD_API_VERSION)
	urlStr := getFullAPIURL(ETCD_CLIENT_PORT, path)

	headers := map[string]string{
		"Content-Type": "application/json",
	}
	p := goutils.HttpRequestParams{
		HttpRequestType: "GET",
		Url:             urlStr,
		Headers:         headers,
	}

	_, jsonResponse := goutils.HttpCreateRequest(p)
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

	headers := map[string]string{
		"Content-Type":   "application/x-www-form-urlencoded",
		"Content-Length": strconv.Itoa(len(data)),
	}

	p := goutils.HttpRequestParams{
		HttpRequestType: "PUT",
		Url:             urlStr,
		Data:            data,
		Headers:         headers,
	}
	goutils.HttpCreateRequest(p)
}

func waitForMetadata(
	resultNode *ResultNode,
	fleetMachine *FleetMachine,
) {

	// Issue request to get machines & parse it. Sleep if cluster not ready yet
	id := strings.Split(resultNode.Key, "fleet/machines/")[1]
	path := fmt.Sprintf(
		"%s/keys/_coreos.com/fleet/machines/%s/object", ETCD_API_VERSION, id)

	url := getFullAPIURL(ETCD_CLIENT_PORT, path)

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	p := goutils.HttpRequestParams{
		HttpRequestType: "GET",
		Url:             url,
		Headers:         headers,
	}

	_, jsonResponse := goutils.HttpCreateRequest(p)

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

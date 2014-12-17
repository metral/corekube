package lib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/coreos/fleet/unit"
)

var FLEET_API_VERSION string = "v1-alpha"
var FLEET_API_PORT string = "10001"

// Types Result, ResultNode, NodeResult & Node adapted from:
// https://github.com/coreos/fleet/blob/master/etcd/result.go

type Map map[string]interface{}

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

type FleetMachines []FleetMachine
type FleetMachine struct {
	ID             string
	PublicIP       string
	Metadata       Map
	Version        string
	TotalResources Map
}

type FleetUnitState struct {
	Hash               string
	MachineID          string
	Name               string
	SystemdActiveState string
	SystemdLoadState   string
	SystemdSubState    string
}

type FleetUnitStates struct {
	States []FleetUnitState
}

type Conf struct {
	BinariesURL string `json:"binariesURL"`
}

func (f FleetMachine) PrintString() {
	log.Printf("-- ID: %s\n", f.ID)
	log.Printf("-- IP: %s\n", f.PublicIP)
	log.Printf("-- Metadata: %s\n", f.Metadata.String())
}

func (m Map) String() string {
	output := ""
	for k, v := range m {
		output += fmt.Sprintf("(%s => %s) ", k, v)
	}
	return output
}

func lowerCasingOfUnitOptionsStr(json_str string) string {
	json_str = strings.Replace(json_str, "Section", "section", -1)
	json_str = strings.Replace(json_str, "Name", "name", -1)
	json_str = strings.Replace(json_str, "Value", "value", -1)

	return json_str
}

func getUnitPathInfo() []map[string]string {
	templatePath := "/units/kubernetes_units"
	unitPathInfo := []map[string]string{}

	unitPathInfo = append(unitPathInfo,
		map[string]string{
			"path":     templatePath + "/download",
			"subState": "exited",
		},
	)

	unitPathInfo = append(unitPathInfo,
		map[string]string{
			"path":     templatePath + "/roles",
			"subState": "running",
		},
	)

	return unitPathInfo
}

func getSubStateByPath(path string) string {
	unitPathInfo := getUnitPathInfo()

	for _, v := range unitPathInfo {
		if v["path"] == path {
			return v["subState"]
		}
	}

	return ""
}

func createUnitFiles(fleetMachine *FleetMachine) []string {
	unitPathInfo := getUnitPathInfo()
	createdFiles := []string{}

	perm := os.FileMode(os.ModeDir)

	for _, v := range unitPathInfo {
		err := os.RemoveAll(v["path"])
		checkForErrors(err)

		os.MkdirAll(v["path"], perm)
	}

	switch fleetMachine.Metadata["kubernetes_role"] {
	case "master":
		createdFiles = createMasterUnits(fleetMachine, unitPathInfo)
	case "minion":
		createdFiles = createMinionUnits(fleetMachine, unitPathInfo)
	}

	log.Printf("Created all unit files for: %s\n", fleetMachine.ID)
	return createdFiles
}

func createMasterUnits(
	fleetMachine *FleetMachine,
	unitPathInfo []map[string]string,
) []string {

	files := map[string]string{
		"api":        "master-apiserver@.service",
		"controller": "master-controller-manager@.service",
		"scheduler":  "master-scheduler@.service",
		"download":   "master-download-kubernetes@.service",
	}

	createdFiles := []string{}

	// Form download service file from template
	readfile, err := ioutil.ReadFile(
		fmt.Sprintf("/templates/%s", files["download"]))
	checkForErrors(err)
	download := string(readfile)
	download = strings.Replace(download, "<ID>", fleetMachine.ID, -1)

	file, _ := os.Open("/templates/conf.json")
	conf := new(Conf)
	json.NewDecoder(file).Decode(conf)
	download = strings.Replace(download, "<URL>", conf.BinariesURL, -1)

	// Write download service file
	filename := strings.Replace(files["download"], "@", "@"+fleetMachine.ID, -1)
	download_file := fmt.Sprintf("%s/%s",
		unitPathInfo[0]["path"], filename)
	err = ioutil.WriteFile(download_file, []byte(download), 0644)
	checkForErrors(err)
	createdFiles = append(createdFiles, download_file)

	// Form apiserver service file from template
	readfile, err = ioutil.ReadFile(
		fmt.Sprintf("/templates/%s", files["api"]))
	checkForErrors(err)
	apiserver := string(readfile)
	apiserver = strings.Replace(apiserver, "<ID>", fleetMachine.ID, -1)

	// Write apiserver service file
	filename = strings.Replace(files["api"], "@", "@"+fleetMachine.ID, -1)
	apiserver_file := fmt.Sprintf("%s/%s", unitPathInfo[1]["path"], filename)
	err = ioutil.WriteFile(apiserver_file, []byte(apiserver), 0644)
	checkForErrors(err)
	createdFiles = append(createdFiles, apiserver_file)

	// Form controller service file from template
	readfile, err = ioutil.ReadFile(
		fmt.Sprintf("/templates/%s", files["controller"]))
	checkForErrors(err)
	controller := string(readfile)
	controller = strings.Replace(controller, "<ID>", fleetMachine.ID, -1)

	// Write controller service file
	filename = strings.Replace(files["controller"], "@", "@"+fleetMachine.ID, -1)
	controller_file := fmt.Sprintf("%s/%s", unitPathInfo[1]["path"], filename)
	err = ioutil.WriteFile(controller_file, []byte(controller), 0644)
	checkForErrors(err)
	createdFiles = append(createdFiles, controller_file)

	// Form scheduler service file from template
	readfile, err = ioutil.ReadFile(
		fmt.Sprintf("/templates/%s", files["scheduler"]))
	checkForErrors(err)
	scheduler := string(readfile)
	scheduler = strings.Replace(scheduler, "<ID>", fleetMachine.ID, -1)

	// Write scheduler service file
	filename = strings.Replace(files["scheduler"], "@", "@"+fleetMachine.ID, -1)
	scheduler_file := fmt.Sprintf("%s/%s", unitPathInfo[1]["path"], filename)
	err = ioutil.WriteFile(scheduler_file, []byte(scheduler), 0644)
	checkForErrors(err)
	createdFiles = append(createdFiles, scheduler_file)

	return createdFiles
}

func createMinionUnits(fleetMachine *FleetMachine,
	unitPathInfo []map[string]string,
) []string {
	files := map[string]string{
		"kubelet":  "minion-kubelet@.service",
		"proxy":    "minion-proxy@.service",
		"download": "minion-download-kubernetes@.service",
	}

	createdFiles := []string{}

	// Form download service file from template
	readfile, err := ioutil.ReadFile(
		fmt.Sprintf("/templates/%s", files["download"]))
	checkForErrors(err)
	download := string(readfile)
	download = strings.Replace(download, "<ID>", fleetMachine.ID, -1)

	file, _ := os.Open("/templates/conf.json")
	conf := new(Conf)
	json.NewDecoder(file).Decode(conf)
	download = strings.Replace(download, "<URL>", conf.BinariesURL, -1)

	// Write download service file
	filename := strings.Replace(files["download"], "@", "@"+fleetMachine.ID, -1)
	download_file := fmt.Sprintf("%s/%s",
		unitPathInfo[0]["path"], filename)
	err = ioutil.WriteFile(download_file, []byte(download), 0644)
	checkForErrors(err)
	createdFiles = append(createdFiles, download_file)

	// Form kubelet service file from template
	readfile, err = ioutil.ReadFile(
		fmt.Sprintf("/templates/%s", files["kubelet"]))
	checkForErrors(err)
	kubelet := string(readfile)
	kubelet = strings.Replace(kubelet, "<ID>", fleetMachine.ID, -1)
	kubelet = strings.Replace(kubelet, "<IP_ADDR>", fleetMachine.PublicIP, -1)

	// Write kubelet service file
	filename = strings.Replace(files["kubelet"], "@", "@"+fleetMachine.ID, -1)
	kubelet_file := fmt.Sprintf("%s/%s", unitPathInfo[1]["path"], filename)
	err = ioutil.WriteFile(kubelet_file, []byte(kubelet), 0644)
	checkForErrors(err)
	createdFiles = append(createdFiles, kubelet_file)

	// Form proxy service file from template
	readfile, err = ioutil.ReadFile(
		fmt.Sprintf("/templates/%s", files["proxy"]))
	checkForErrors(err)
	proxy := string(readfile)
	proxy = strings.Replace(proxy, "<ID>", fleetMachine.ID, -1)

	// Write proxy service file
	filename = strings.Replace(files["proxy"], "@", "@"+fleetMachine.ID, -1)
	proxy_file := fmt.Sprintf("%s/%s", unitPathInfo[1]["path"], filename)
	err = ioutil.WriteFile(proxy_file, []byte(proxy), 0644)
	checkForErrors(err)
	createdFiles = append(createdFiles, proxy_file)

	return createdFiles
}

func startUnitFile(unitFile string) {
	filename := filepath.Base(unitFile)
	unitFilepath := fmt.Sprintf(
		"%s/units/%s", FLEET_API_VERSION, filename)
	url := getFullAPIURL(FLEET_API_PORT, unitFilepath)

	log.Printf("Starting unit file: %s", filename)

	statusCode := 0
	for statusCode != 204 {
		readfile, err := ioutil.ReadFile(unitFile)
		checkForErrors(err)

		content := string(readfile)
		u, _ := unit.NewUnitFile(content)

		options_bytes, _ := json.Marshal(u.Options)
		options_str := lowerCasingOfUnitOptionsStr(string(options_bytes))

		json_str := fmt.Sprintf(
			`{"name": "%s", "desiredState":"launched", "options": %s}`,
			filename,
			options_str)

		resp := httpPutRequest(url, []byte(json_str))
		statusCode = resp.StatusCode

		time.Sleep(1 * time.Second)
		/*
			log.Printf(
				"curl -H \"Content-Type: application/json\" -X PUT "+
					"-d %q localhost:10001/v1-alpha/units/%s",
				json_str, filename)
			body, err := ioutil.ReadAll(resp.Body)
			log.Printf("Status Code: %s", statusCode)
			log.Printf("[Error] in HTTP Body: %s - %v", body, err)
			checkForErrors(err)
		*/
	}
}

func getUnitState(unitFile string) FleetUnitState {
	var fleetUnitStates FleetUnitStates
	filename := filepath.Base(unitFile)

	urlPath := fmt.Sprintf("%s/state", FLEET_API_VERSION)
	url := getFullAPIURL(FLEET_API_PORT, urlPath)

	jsonResponse := httpGetRequest(url)
	err := json.Unmarshal(jsonResponse, &fleetUnitStates)
	checkForErrors(err)

	for _, unitState := range fleetUnitStates.States {
		if unitState.Name == filename {
			return unitState
		}
	}
	return FleetUnitState{}
}

func unitFileCompleted(unitFile string) bool {
	filename := filepath.Base(unitFile)
	dir := filepath.Dir(unitFile)

	expectedSubState := getSubStateByPath(dir)
	unitState := getUnitState(unitFile)

	if unitState.Name == filename {
		if unitState.SystemdSubState == expectedSubState {
			return true
		}
	}

	return false
}

func waitUnitFileComplete(unitFile string) {
	filename := filepath.Base(unitFile)

	complete := false
	for !complete {
		complete = unitFileCompleted(unitFile)

		if !complete {
			log.Printf("-- Waiting for the following unit file to complete: %s",
				filename)
			time.Sleep(1 * time.Second)
		}
	}

	log.Printf("The following unit file has completed: %s", filename)
}

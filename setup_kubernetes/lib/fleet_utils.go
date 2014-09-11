package lib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/coreos/fleet/unit"
)

type Map map[string]interface{}

type FleetMachines struct {
	Action string
	Node   FleetMachinesNode
}

type FleetMachinesNode struct {
	Key           string
	Dir           bool
	Nodes         []FleetMachinesNodeNodesValue
	ModifiedIndex int
	CreatedIndex  int
}

type FleetMachinesNodeNodesValue struct {
	Key           string
	Dir           bool
	ModifiedIndex int
	CreatedIndex  int
}

type FleetMachineObject struct {
	Action string
	Node   FleetMachineObjectNode
}

type FleetMachineObjectNode struct {
	Key           string
	Value         string
	Expiration    string
	Ttl           int
	ModifiedIndex int
	CreatedIndex  int
}

type FleetMachineObjectNodeValue struct {
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

func (f FleetMachinesNodeNodesValue) String() string {
	output := fmt.Sprintf(
		"Key: %s | Dir: %t | ModifiedIndex: %d | CreatedIndex: %d",
		f.Key,
		f.Dir,
		f.ModifiedIndex,
		f.CreatedIndex,
	)
	return output
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

func StartUnitsInDir(path string) {
	files, _ := ioutil.ReadDir(path)

	for _, f := range files {
		statusCode := 0
		for statusCode != 204 {
			unitpath := fmt.Sprintf("v1-alpha/units/%s", f.Name())
			url := getFullAPIURL("10001", unitpath)
			filepath := fmt.Sprintf("%s/%s", path, f.Name())

			readfile, err := ioutil.ReadFile(filepath)
			checkForErrors(err)
			content := string(readfile)

			u, _ := unit.NewUnitFile(content)

			options_bytes, _ := json.Marshal(u.Options)
			options_str := lowerCasingOfUnitOptionsStr(string(options_bytes))

			json_str := fmt.Sprintf(
				`{"name": "%s", "desiredState":"launched", "options": %s}`,
				f.Name(),
				options_str)

			resp := httpPutRequest(url, []byte(json_str))
			statusCode = resp.StatusCode

			if statusCode != 204 {
				time.Sleep(1 * time.Second)
			}
		}
	}
}

func stringInSlice(a string, list []os.FileInfo) bool {
	for _, b := range list {
		if b.Name() == a {
			return true
		}
	}
	return false
}

func CheckUnitsState(path, activeState, subState string) {

	var fleetUnitStates FleetUnitStates

	url := getFullAPIURL("10001", "v1-alpha/state")
	jsonResponse := httpGetRequest(url)
	err := json.Unmarshal(jsonResponse, &fleetUnitStates)
	checkForErrors(err)

	files, _ := ioutil.ReadDir(path)

	totalKubernetesMachines := len(files)
	activeExitedCount := 0
	for activeExitedCount < totalKubernetesMachines {
		for _, unit := range fleetUnitStates.States {
			if stringInSlice(unit.Name, files) &&
				unit.SystemdActiveState == activeState &&
				unit.SystemdSubState == subState {
				activeExitedCount += 1
			}
		}
		if activeExitedCount == totalKubernetesMachines {
			break
		}
		log.Printf("Waiting for (%d) services to be complete "+
			"in fleet. Currently at: (%d)",
			totalKubernetesMachines, activeExitedCount)
		activeExitedCount = 0
		time.Sleep(1 * time.Second)
		jsonResponse := httpGetRequest(url)
		err := json.Unmarshal(jsonResponse, &fleetUnitStates)
		checkForErrors(err)
	}

	log.Printf("Unit files in '%s' have completed", path)
}

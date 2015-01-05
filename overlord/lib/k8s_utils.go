package lib

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/metral/goutils"
)

var K8S_API_VERSION string = "v1beta1"
var K8S_API_PORT string = "8080"

type PreregisteredMinion struct {
	Kind       string `json:"kind,omitempty"`
	ID         string `json:"id,omitempty"`
	HostIP     string `json:"hostIP,omitempty"`
	APIVersion string `json:"apiVersion,omitempty"`
}

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

func registerMinions(master *FleetMachine, minions *FleetMachines) {

	// Get registered minions, if any
	endpoint := fmt.Sprintf("http://%s:%s", master.PublicIP, K8S_API_PORT)
	masterAPIurl := fmt.Sprintf("%s/api/%s/minions", endpoint, K8S_API_VERSION)

	headers := map[string]string{
		"Content-Type": "application/json",
	}
	p := goutils.HttpRequestParams{
		HttpRequestType: "GET",
		Url:             masterAPIurl,
		Headers:         headers,
	}

	_, jsonResponse := goutils.HttpCreateRequest(p)
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

func register(endpoint, addr string) error {
	m := &PreregisteredMinion{
		Kind:       "Minion",
		APIVersion: K8S_API_VERSION,
		ID:         addr,
		HostIP:     addr,
	}
	data, err := json.Marshal(m)
	goutils.CheckForErrors(goutils.ErrorParams{Err: err, CallerNum: 1})

	url := fmt.Sprintf("%s/api/%s/minions", endpoint, K8S_API_VERSION)

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	p := goutils.HttpRequestParams{
		HttpRequestType: "POST",
		Url:             url,
		Data:            data,
		Headers:         headers,
	}
	statusCode, _ := goutils.HttpCreateRequest(p)

	switch statusCode {
	case 200, 202:
		log.Printf("------------------------------------------------")
		log.Printf("Registered machine with the master: %s\n", addr)
		return nil
	case 409:
		return nil
	}
	return nil
}

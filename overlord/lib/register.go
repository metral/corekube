package lib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var K8S_API_VERSION string = "v1beta1"
var K8S_API_PORT string = "8080"

type PreregisteredMinion struct {
	Kind       string `json:"kind,omitempty"`
	ID         string `json:"id,omitempty"`
	HostIP     string `json:"hostIP,omitempty"`
	APIVersion string `json:"apiVersion,omitempty"`
}

func register(endpoint, addr string) error {
	m := &PreregisteredMinion{
		Kind:       "Minion",
		APIVersion: K8S_API_VERSION,
		ID:         addr,
		HostIP:     addr,
	}
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/api/%s/minions", endpoint, K8S_API_VERSION)
	res, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case 200, 202:
		log.Printf("------------------------------------------------")
		log.Printf("Registered machine with the master: %s\n", addr)
		return nil
	case 409:
		//log.Printf("Machine has already registered with master: %s\n", addr)
		return nil
	}
	data, err = ioutil.ReadAll(res.Body)
	log.Printf("Response: %#v", res)
	log.Printf("Response Body:\n%s", string(data))
	return errors.New("error registering: " + addr)
}

package lib

import "fmt"

type EtcdAdminMachine struct {
	Name      string
	State     string
	ClientURL string
	PeerURL   string
}

type EtcdAdminMachines []EtcdAdminMachine

// Retrieve string of EtcdMachine
func (e EtcdAdminMachine) String() string {
	output := fmt.Sprintf("Name: %s | State: %s | ClientURL: %s | PeerURL: %s",
		e.Name,
		e.State,
		e.ClientURL,
		e.PeerURL,
	)
	return output
}

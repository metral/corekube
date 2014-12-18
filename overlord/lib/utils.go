package lib

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/metral/goutils"
)

// Get the IP address of the docker host as this is run from within container
func getDockerHostIP() string {
	cmd := fmt.Sprintf("netstat -nr | grep '^0\\.0\\.0\\.0' | awk '{print $2}'")
	out, err := exec.Command("sh", "-c", cmd).Output()
	goutils.CheckForErrors(goutils.ErrorParams{Err: err, CallerNum: 1})

	ip := string(out)
	ip = strings.Replace(ip, "\n", "", -1)
	return ip
}

func removeOverlord(nodes *ResultNodes) {
	var fleetMachine FleetMachine
	n := *nodes

	for i, node := range n {
		waitForMetadata(&node, &fleetMachine)
		if fleetMachine.Metadata["kubernetes_role"] == "overlord" {
			n = append(n[:i], n[i+1:]...)
			*nodes = n
			break
		}
	}

}

func Main() {
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
			waitForMetadata(&resultNode, &fleetMachine)

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

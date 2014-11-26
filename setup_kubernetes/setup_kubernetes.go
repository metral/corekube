package main

import (
	"fmt"
	"log"
	"os"
	"setup_kubernetes/lib"
)

// Access the CoreOS / docker etcd API to extract machine information
func main() {
	masterCount, minionCount := lib.SetupFlags()
	expectedMachineCount := masterCount + minionCount

	if expectedMachineCount <= 0 {
		lib.Usage()
		os.Exit(2)
	}

	// Get fleet machines & metadata
	var fleetMachinesAbstract lib.FleetMachinesAbstract
	lib.WaitForFleetMachines(&fleetMachinesAbstract, expectedMachineCount)

	var fleetMachines []lib.FleetMachine
	for _, value := range fleetMachinesAbstract.Node.Nodes {

		// Get fleet metadata
		var fleetMachine lib.FleetMachine
		lib.WaitForFleetMachineMetadata(
			&value,
			&fleetMachine,
			expectedMachineCount)

		fleetMachines = append(
			fleetMachines, fleetMachine)
		log.Printf(
			"\nFleet Machine:\n-- ID: %s\n-- PublicIP: %s\n-- Metadata: %s\n\n",
			fleetMachine.ID,
			fleetMachine.PublicIP,
			fleetMachine.Metadata.String(),
		)
	}

	// Create all systemd unit files from templates
	path := "/units/kubernetes_units"

	// Start all systemd unit files in specified path via fleet
	unitPathInfo := []map[string]string{}
	unitPathInfo = append(unitPathInfo, map[string]string{
		"path":        path + "/download",
		"activeState": "active", "subState": "exited"})
	unitPathInfo = append(unitPathInfo, map[string]string{
		"path":        path + "/roles",
		"activeState": "active", "subState": "running"})

	lib.CreateUnitFiles(&fleetMachines, unitPathInfo)

	// Start & check state for download & role units
	for _, v := range unitPathInfo {
		lib.StartUnitsInDir(v["path"])
		lib.CheckUnitsState(v["path"], v["activeState"], v["subState"])
	}

	// Register minions with master
	masterIP := lib.FindInfoForRole("master", &fleetMachines)[0]
	minionIPs := lib.FindInfoForRole("minion", &fleetMachines)
	k8sAPI := fmt.Sprintf("http://%s:8080", masterIP)
	for _, minionIP := range minionIPs {
		lib.Register(k8sAPI, minionIP)
	}
}

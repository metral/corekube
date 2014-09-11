package main

import (
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
	var fleetMachines lib.FleetMachines
	lib.WaitForFleetMachines(&fleetMachines, expectedMachineCount)

	var fleetMachineEntities []lib.FleetMachineObjectNodeValue
	for _, fleetMachinesNodeNodesValue := range fleetMachines.Node.Nodes {

		// Get fleet metadata
		var fleetMachineObjectNodeValue lib.FleetMachineObjectNodeValue
		lib.WaitForFleetMachineMetadata(
			&fleetMachinesNodeNodesValue,
			&fleetMachineObjectNodeValue,
			expectedMachineCount)

		fleetMachineEntities = append(
			fleetMachineEntities, fleetMachineObjectNodeValue)
		log.Printf(
			"\nFleet Machine:\n-- ID: %s\n-- PublicIP: %s\n-- Metadata: %s\n\n",
			fleetMachineObjectNodeValue.ID,
			fleetMachineObjectNodeValue.PublicIP,
			fleetMachineObjectNodeValue.Metadata.String(),
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

	lib.CreateUnitFiles(&fleetMachineEntities, unitPathInfo)
	for _, v := range unitPathInfo {
		lib.StartUnitsInDir(v["path"])
		lib.CheckUnitsState(v["path"], v["activeState"], v["subState"])
	}
}

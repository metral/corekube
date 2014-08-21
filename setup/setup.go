package main

import (
	"log"
	"os"
	"setup/lib"
)

// Access the CoreOS / docker etcd API to extract machine information
func main() {
	expectedMachineCount := lib.SetupFlags()

	if expectedMachineCount <= 0 {
		lib.Usage()
		os.Exit(2)
	}

	// Get fleet machines & metadata
	var fleetMachines lib.FleetMachines
	lib.WaitForFleetMachines(&fleetMachines, expectedMachineCount)

	// TODO: delete
	for _, fleetMachinesNodeNodesValue := range fleetMachines.Node.Nodes {

		// Get fleet metadata
		var fleetMachineObjectNodeValue lib.FleetMachineObjectNodeValue
		lib.WaitForFleetMachineMetadata(
			&fleetMachinesNodeNodesValue,
			&fleetMachineObjectNodeValue,
			expectedMachineCount)

		log.Printf(
			"\nFleet Machine:\n-- ID: %s\n-- PublicIP: %s\n-- Metadata: %s\n\n",
			fleetMachineObjectNodeValue.ID,
			fleetMachineObjectNodeValue.PublicIP,
			fleetMachineObjectNodeValue.Metadata.String(),
		)
	}
}

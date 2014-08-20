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
		log.Printf("%s\n", fleetMachinesNodeNodesValue.String())

		// Get fleet metadata
		var fleetMachineObjectNodeValue lib.FleetMachineObjectNodeValue
		lib.WaitForFleetMachineMetadata(
			&fleetMachinesNodeNodesValue,
			&fleetMachineObjectNodeValue,
			expectedMachineCount)
		log.Printf("%s",
			fleetMachineObjectNodeValue.Metadata["kubernetes_role"])

		// TODO wait until all 3 have metadata
		/*
			// TODO: delete
			for _, node := range fleetMachines.Node.Nodes {
				log.Printf("%s\n", node.String())
			}
		*/
	}

}

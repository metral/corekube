package lib

import "fmt"

type FleetMachines struct {
	Action string
	Node   FleetMachinesNode
}

type FleetMachinesNode struct {
	Key           string
	Dir           bool
	Nodes         []FleetMachinesNodeNodesValue
	ModifiedIndex uint64
	CreatedIndex  uint64
}

type FleetMachinesNodeNodesValue struct {
	Key           string
	Dir           bool
	ModifiedIndex uint64
	CreatedIndex  uint64
}

type FleetMachineObject struct {
	Action string
	Node   FleetMachineObjectNode
}

type FleetMachineObjectNode struct {
	Key           string
	Value         string
	Expiration    string
	Ttl           uint64
	ModifiedIndex uint64
	CreatedIndex  uint64
}

type FleetMachineObjectNodeValue struct {
	ID             string
	PublicIP       string
	Metadata       map[string]interface{}
	Version        string
	TotalResources map[string]interface{}
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

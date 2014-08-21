package lib

import "fmt"

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

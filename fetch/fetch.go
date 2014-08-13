package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/coreos/go-etcd/etcd"
)

type NodeGroup []*etcd.Node //NodeGroup is a slice of pointers to etcd Nodes

// Sort Interface implementation methods
func (n NodeGroup) Len() int {
	return len(n)
}

func (n NodeGroup) Less(i, j int) bool {
	if n[i].Key < n[j].Key {
		return true
	}
	return false
}

func (n NodeGroup) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

func Usage() {
	fmt.Printf("Usage: %s\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

func SetupFlags() (discoveryHost, discoveryPath *string) {
	discoveryHost = flag.String("discovery_host",
		"http://127.0.0.1:4001", "Discovery URL:Port")
	discoveryPath = flag.String("discovery_path",
		"",
		"Discovery path i.e. _etcd/registry/uVa2GHOTTxl27eyKk6clBwyaurf7KiWd")

	flag.Parse()

	if *discoveryHost == "" || *discoveryPath == "" {
		Usage()
	}

	return discoveryHost, discoveryPath
}

func main() {
	// Connect to the etcd discovery to pull the nodes
	discoveryHost, discoveryPath := SetupFlags()

	client := etcd.NewClient([]string{*discoveryHost})
	resp, _ := client.Get(*discoveryPath, false, false)

	// Store the pointer to the etcd nodes as a NodeGroup
	group := NodeGroup{}
	for _, n := range resp.Node.Nodes {
		group = append(group, n)
	}

	// Sort the NodeGroup
	sort.Sort(group)

	// Print out sorted NodeGroup by key
	for _, n := range group {
		log.Printf("%s: %s\n", n.Key, n.Value)
	}
}

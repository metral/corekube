package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/coreos/go-etcd/etcd"
)

func Usage() {
	fmt.Printf("Usage: %s\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

func SetupFlags() (discoveryHost, discoveryPath *string) {
	discoveryURL := flag.String("discovery_url", "", "Discovery URL")
	flag.Parse()

	u, err := url.Parse(*discoveryURL)
	if err != nil {
		panic(err)
	}

	discoveryHost = new(string)
	*discoveryHost = u.Scheme + "://" + u.Host

	path := strings.Split(u.Path, "/keys/")[1]
	discoveryPath = new(string)
	*discoveryPath = path

	if *discoveryHost == "" || *discoveryPath == "" {
		Usage()
	}

	return discoveryHost, discoveryPath
}

func main() {
	// Connect to the etcd discovery to pull the nodes
	discoveryHost, discoveryPath := SetupFlags()

	client := etcd.NewClient([]string{*discoveryHost})
	resp, _ := client.Get(*discoveryPath, true, false)

	// Store the pointer to the etcd nodes as a NodeGroup
	for _, n := range resp.Node.Nodes {
		log.Printf("%s: %s\n", n.Key, n.Value)
	}
}

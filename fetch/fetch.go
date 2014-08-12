package main

import (
	"log"

	"github.com/coreos/go-etcd/etcd"
)

func main() {
	client := etcd.NewClient([]string{"http://104.130.8.142:4001"})
	resp, err := client.Get("testcluster", false, false)
	if err != nil {
		log.Fatal(err)
	}
	for _, n := range resp.Node.Nodes {
		log.Printf("%s: %s\n", n.Key, n.Value)
	}
}

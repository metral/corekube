package main

import (
	"fmt"
	"log"
	"net/url"
	"os/exec"
	"strings"

	"github.com/coreos/go-etcd/etcd"
)

func ParseDiscovery() (discoveryHost, discoveryPath *string) {
	// Pull the ETCD_DISCOVERY env var and parse it for the etcd client usage
	file := "20-cloudinit.conf"
	cmd := fmt.Sprintf("cat %s | grep ETCD_DISCOVERY | cut -d '=' -f 3 | cut -d '\"' -f 1", file)
	out, err := exec.Command("sh", "-c", cmd).Output()

	discoveryURL := string(out)

	u, err := url.Parse(discoveryURL)
	if err != nil {
		panic(err)
	}

	discoveryHost = new(string)
	*discoveryHost = u.Scheme + "://" + u.Host

	path := strings.Split(u.Path, "/keys/")[1]
	discoveryPath = new(string)
	*discoveryPath = path

	return discoveryHost, discoveryPath
}

func main() {
	discoveryHost, discoveryPath := ParseDiscovery()

	// Connect to the etcd discovery to pull the nodes
	client := etcd.NewClient([]string{*discoveryHost})
	resp, _ := client.Get(*discoveryPath, true, false)

	for _, n := range resp.Node.Nodes {
		log.Printf("%s: %s\n", n.Key, n.Value)
	}
}

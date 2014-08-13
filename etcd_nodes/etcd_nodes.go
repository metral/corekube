package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/coreos/go-etcd/etcd"
)

func ParseDiscovery() (discoveryHost, discoveryPath string) {
	// Pull the ETCD_DISCOVERY env var and parse it for the etcd client usage
	file := "/mnt/20-cloudinit.conf"
	cmd := fmt.Sprintf("cat %s | grep ETCD_DISCOVERY | cut -d '=' -f 3 | cut -d '\"' -f 1", file)
	out, err := exec.Command("sh", "-c", cmd).Output()

	discoveryURL := string(out)

	u, err := url.Parse(discoveryURL)
	if err != nil {
		panic(err)
	}

	discoveryHost = fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	path := strings.Split(u.Path, "/keys/")[1]
	path = strings.Replace(path, "\n", "", -1)
	discoveryPath = fmt.Sprintf("%s", path)

	return discoveryHost, discoveryPath
}

func main() {
	discoveryHost, discoveryPath := ParseDiscovery()

	// Connect to the etcd discovery to pull the nodes
	client := etcd.NewClient([]string{discoveryHost})
	resp, err := client.Get(discoveryPath, true, false)

	if err != nil {
		log.Printf("%s", err)
		os.Exit(2)
	}

	for _, n := range resp.Node.Nodes {
		log.Printf("%s: %s\n", n.Key, n.Value)
	}
}

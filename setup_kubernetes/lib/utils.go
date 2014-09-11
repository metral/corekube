package lib

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Check for errors and panic, if found
func checkForErrors(err error) {
	if err != nil {
		pc, fn, line, _ := runtime.Caller(1)
		msg := fmt.Sprintf("[Error] in %s[%s:%d] %v",
			runtime.FuncForPC(pc).Name(), fn, line, err)
		log.Fatal(msg)
	}
}

// Get the IP address of the docker host as this is run from within container
func getDockerHostIP() string {
	cmd := fmt.Sprintf("netstat -nr | grep '^0\\.0\\.0\\.0' | awk '{print $2}'")
	out, err := exec.Command("sh", "-c", cmd).Output()
	checkForErrors(err)

	ip := string(out)
	ip = strings.Replace(ip, "\n", "", -1)
	return ip
}

// Compose the etcd API host:port location
func getEtcdAPI(host string, port string) string {
	return fmt.Sprintf("http://%s:%s", host, port)
}

func httpGetRequest(url string) []byte {
	resp, err := http.Get(url)
	checkForErrors(err)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	checkForErrors(err)

	return body
}

func httpPutRequest(url string, json_data []byte) *http.Response {
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(json_data))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	checkForErrors(err)

	defer resp.Body.Close()

	return resp
}

func getFullAPIURL(port, etcdAPIPath string) string {
	etcdAPI := getEtcdAPI(getDockerHostIP(), port)
	url := fmt.Sprintf("%s/%s", etcdAPI, etcdAPIPath)
	return url
}

func WaitForFleetMachines(
	fleetMachines *FleetMachines, expectedMachineCount int) {

	// Issue request to get machines & parse it. Sleep if cluster not ready yet
	url := getFullAPIURL("4001", "v2/keys/_coreos.com/fleet/machines")
	jsonResponse := httpGetRequest(url)
	err := json.Unmarshal(jsonResponse, fleetMachines)
	checkForErrors(err)
	totalMachines := len(fleetMachines.Node.Nodes)

	for totalMachines < expectedMachineCount {
		log.Printf("Waiting for all (%d) machines to be available "+
			"in fleet. Currently at: (%d)",
			expectedMachineCount, totalMachines)
		time.Sleep(1 * time.Second)

		jsonResponse := httpGetRequest(url)
		err := json.Unmarshal(jsonResponse, fleetMachines)
		checkForErrors(err)
		totalMachines = len(fleetMachines.Node.Nodes)
	}
}

func WaitForFleetMachineMetadata(
	fleetMachinesNodeNodesValue *FleetMachinesNodeNodesValue,
	fleetMachineObjectNodeValue *FleetMachineObjectNodeValue,
	expectedMachineCount int) {

	// Issue request to get machines & parse it. Sleep if cluster not ready yet
	id := strings.Split(fleetMachinesNodeNodesValue.Key, "fleet/machines/")[1]
	path := fmt.Sprintf("v2/keys/_coreos.com/fleet/machines/%s/object", id)

	url := getFullAPIURL("4001", path)
	jsonResponse := httpGetRequest(url)

	var fleetMachineObject FleetMachineObject
	err := json.Unmarshal(jsonResponse, &fleetMachineObject)
	checkForErrors(err)

	err = json.Unmarshal(
		[]byte(fleetMachineObject.Node.Value), &fleetMachineObjectNodeValue)
	checkForErrors(err)

	for len(fleetMachineObjectNodeValue.Metadata) == 0 ||
		fleetMachineObjectNodeValue.Metadata["kubernetes_role"] == nil {
		log.Printf("Waiting for machine (%s) metadata to be available "+
			"in fleet...", fleetMachineObjectNodeValue.ID)
		time.Sleep(1 * time.Second)

		err = json.Unmarshal(
			[]byte(fleetMachineObject.Node.Value), &fleetMachineObjectNodeValue)
		checkForErrors(err)

	}
}

func createMasterUnits(
	entity *FleetMachineObjectNodeValue,
	minionIPAddrs string,
	unitPathInfo []map[string]string,
) {

	files := map[string]string{
		"api":        "master-apiserver@.service",
		"controller": "master-controller-manager@.service",
		"download":   "master-download-kubernetes@.service",
	}

	// Form apiserver service file from template
	readfile, err := ioutil.ReadFile(
		fmt.Sprintf("/templates/%s", files["api"]))
	checkForErrors(err)
	apiserver := string(readfile)
	apiserver = strings.Replace(apiserver, "<ID>", entity.ID, -1)
	apiserver = strings.Replace(
		apiserver, "<MINION_IP_ADDRS>", minionIPAddrs, -1)

	// Write apiserver service file
	filename := strings.Replace(files["api"], "@", "@"+entity.ID, -1)
	apiserver_file := fmt.Sprintf("%s/%s", unitPathInfo[1]["path"], filename)
	err = ioutil.WriteFile(apiserver_file, []byte(apiserver), 0644)
	checkForErrors(err)

	// Form controller service file from template
	readfile, err = ioutil.ReadFile(
		fmt.Sprintf("/templates/%s", files["controller"]))
	checkForErrors(err)
	controller := string(readfile)
	controller = strings.Replace(controller, "<ID>", entity.ID, -1)

	// Write controller service file
	filename = strings.Replace(files["controller"], "@", "@"+entity.ID, -1)
	controller_file := fmt.Sprintf("%s/%s", unitPathInfo[1]["path"], filename)
	err = ioutil.WriteFile(controller_file, []byte(controller), 0644)
	checkForErrors(err)

	// Form download service file from template
	readfile, err = ioutil.ReadFile(
		fmt.Sprintf("/templates/%s", files["download"]))
	checkForErrors(err)
	download := string(readfile)
	download = strings.Replace(download, "<ID>", entity.ID, -1)

	// Write download service file
	filename = strings.Replace(files["download"], "@", "@"+entity.ID, -1)
	download_file := fmt.Sprintf("%s/%s",
		unitPathInfo[0]["path"], filename)
	err = ioutil.WriteFile(download_file, []byte(download), 0644)
	checkForErrors(err)
}

func createMinionUnits(entity *FleetMachineObjectNodeValue,
	unitPathInfo []map[string]string,
) {
	files := map[string]string{
		"kubelet":  "minion-kubelet@.service",
		"proxy":    "minion-proxy@.service",
		"download": "minion-download-kubernetes@.service",
	}

	// Form kubelet service file from template
	readfile, err := ioutil.ReadFile(
		fmt.Sprintf("/templates/%s", files["kubelet"]))
	checkForErrors(err)
	kubelet := string(readfile)
	kubelet = strings.Replace(kubelet, "<ID>", entity.ID, -1)
	kubelet = strings.Replace(kubelet, "<IP_ADDR>", entity.PublicIP, -1)

	// Write kubelet service file
	filename := strings.Replace(files["kubelet"], "@", "@"+entity.ID, -1)
	kubelet_file := fmt.Sprintf("%s/%s", unitPathInfo[1]["path"], filename)
	err = ioutil.WriteFile(kubelet_file, []byte(kubelet), 0644)
	checkForErrors(err)

	// Form proxy service file from template
	readfile, err = ioutil.ReadFile(
		fmt.Sprintf("/templates/%s", files["proxy"]))
	checkForErrors(err)
	proxy := string(readfile)
	proxy = strings.Replace(proxy, "<ID>", entity.ID, -1)

	// Write proxy service file
	filename = strings.Replace(files["proxy"], "@", "@"+entity.ID, -1)
	proxy_file := fmt.Sprintf("%s/%s", unitPathInfo[1]["path"], filename)
	err = ioutil.WriteFile(proxy_file, []byte(proxy), 0644)
	checkForErrors(err)

	// Form download service file from template
	readfile, err = ioutil.ReadFile(
		fmt.Sprintf("/templates/%s", files["download"]))
	checkForErrors(err)
	download := string(readfile)
	download = strings.Replace(download, "<ID>", entity.ID, -1)

	// Write download service file
	filename = strings.Replace(files["download"], "@", "@"+entity.ID, -1)
	download_file := fmt.Sprintf("%s/%s",
		unitPathInfo[0]["path"], filename)
	err = ioutil.WriteFile(download_file, []byte(download), 0644)
	checkForErrors(err)
}

func getMinionIPAddrs(
	fleetMachineEntities *[]FleetMachineObjectNodeValue) string {
	output := ""

	for _, entity := range *fleetMachineEntities {
		switch entity.Metadata["kubernetes_role"] {
		case "minion":
			output += entity.PublicIP + ","
		}
	}

	k := strings.LastIndex(output, ",")
	return output[:k]
}

func CreateUnitFiles(
	fleetMachineEntities *[]FleetMachineObjectNodeValue,
	unitPathInfo []map[string]string,
) {

	perm := os.FileMode(os.ModeDir)

	for _, v := range unitPathInfo {
		err := os.RemoveAll(v["path"])
		checkForErrors(err)

		os.MkdirAll(v["path"], perm)
	}

	for _, entity := range *fleetMachineEntities {
		switch entity.Metadata["kubernetes_role"] {
		case "master":
			minionIPAddrs := getMinionIPAddrs(fleetMachineEntities)
			createMasterUnits(&entity, minionIPAddrs, unitPathInfo)
		case "minion":
			createMinionUnits(&entity, unitPathInfo)
		}
	}
	log.Printf("Created systemd unit files for kubernetes deployment")
}

func Usage() {
	fmt.Printf("Usage: %s\n", os.Args[0])
	flag.PrintDefaults()
}

func SetupFlags() (int, int) {
	masterCount :=
		flag.Int("master_count", 1,
			"Expected number of kubernetes masters in cluster")
	minionCount :=
		flag.Int("minion_count", 2,
			"Expected number of kubernetes minions in cluster")

	flag.Parse()

	return *masterCount, *minionCount
}

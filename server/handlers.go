/*
 * Copyright (C) 2017 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"launchpad.net/wifi-connect/netman"
	"launchpad.net/wifi-connect/utils"
	"launchpad.net/wifi-connect/wifiap"
)

const (
	managementTemplatePath  = "/templates/management.html"
	connectingTemplatePath  = "/templates/connecting.html"
	operationalTemplatePath = "/templates/operational.html"
)

// ResourcesPath absolute path to web static resources
var ResourcesPath = filepath.Join(os.Getenv("SNAP"), "static")

// Data interface representing any data included in a template
type Data interface{}

// SsidsData dynamic data to fulfill the SSIDs page template
type SsidsData struct {
	Ssids []string
}

// ConnectingData dynamic data to fulfill the connect result page template
type ConnectingData struct {
	Ssid string
}

func execTemplate(w http.ResponseWriter, templatePath string, data Data) {
	templateAbsPath := filepath.Join(ResourcesPath, templatePath)
	t, err := template.ParseFiles(templateAbsPath)
	if err != nil {
		fmt.Printf("== wifi-connect/handler: Error loading the template at %v : %v\n", templatePath, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		fmt.Printf("== wifi-connect/handler: Error executing the template at %v : %v\n", templatePath, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// ManagementHandler lists the current available SSIDs
func ManagementHandler(w http.ResponseWriter, r *http.Request) {
	// daemon stores current available ssids in a file
	ssids, err := utils.ReadSsidsFile()
	if err != nil {
		fmt.Printf("== wifi-connect/handler: Error reading SSIDs file: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := SsidsData{Ssids: ssids}

	// parse template
	execTemplate(w, managementTemplatePath, data)
}

// ConnectHandler reads form got ssid and password and tries to connect to that network
func ConnectHandler(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	ssids := r.Form["ssid"]
	if len(ssids) == 0 {
		fmt.Println("== wifi-connect/handler: SSID not available")
		return
	}
	ssid := ssids[0]

	data := ConnectingData{ssid}
	execTemplate(w, connectingTemplatePath, data)

	pwd := ""
	pwds := r.Form["pwd"]
	if len(pwds) > 0 {
		pwd = pwds[0]
	}

	fmt.Printf("== wifi-connect/handler: Connecting to %v\n", ssid)

	cw := wifiap.DefaultClient()
	cw.Disable()

	//connect
	c := netman.DefaultClient()
	c.SetIfaceManaged("wlan0", true, c.GetWifiDevices(c.GetDevices()))
	_, ap2device, ssid2ap := c.Ssids()

	err := c.ConnectAp(ssid, pwd, ap2device, ssid2ap)

	//TODO signal user in portal on failure to connect
	if err != nil {
		fmt.Printf("== wifi-connect/handler: Failed connecting to %v.\n", ssid)
	}

	//remove flag file so that daemon starts checking state
	//and takes control again
	waitPath := os.Getenv("SNAP_COMMON") + "/startingApConnect"
	utils.RemoveFlagFile(waitPath)
}

type disconnectData struct {
}

// OperationalHandler display Opertational mode page
func OperationalHandler(w http.ResponseWriter, r *http.Request) {
	data := disconnectData{}
	execTemplate(w, operationalTemplatePath, data)
}

type hashResponse struct {
	Err       string
	HashMatch bool
}

// HashItHandler returns a hash of the password as json
func HashItHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	hashMe := r.Form["Hash"]
	hashed, errH := utils.MatchingHash(hashMe[0])
	if errH != nil {
		fmt.Println("== wifi-connect/HashitHandler: error hashing:", errH)
		return
	}
	res := &hashResponse{}
	res.HashMatch = hashed
	res.Err = "no error"
	b, err := json.Marshal(res)
	if err != nil {
		fmt.Println("== wifi-connect/HashItHandler: error mashaling json")
		return
	}
	w.Write(b)
}

// DisconnectHandler allows user to disconnect from external AP
func DisconnectHandler(w http.ResponseWriter, r *http.Request) {
	c := netman.DefaultClient()
	c.DisconnectWifi(c.GetWifiDevices(c.GetDevices()))
}

// scanSsids sets wlan0 to be managed and then scans
// for ssids. If found, write the ssids (comma separated)
// to path and return true, else return false.
func scanSsids(path string, c *netman.Client) bool {
	manage(c)
	SSIDs, _, _ := c.Ssids()
	//only write SSIDs when found
	if len(SSIDs) > 0 {
		var out string
		for _, ssid := range SSIDs {
			out += strings.TrimSpace(ssid.Ssid) + ","
		}
		out = out[:len(out)-1]
		err := ioutil.WriteFile(path, []byte(out), 0644)
		if err != nil {
			fmt.Println("== wifi-connect: Error writing SSID(s) to ", path)
		} else {
			fmt.Println("== wifi-connect: SSID(s) obtained")
			return true
		}
	}
	fmt.Println("== wifi-connect: No SSID found")
	return false
}

// unmanage sets wlan0 to be unmanaged by network manager if it
// is managed
func unmanage(c *netman.Client) {
	ifaces, _ := c.WifisManaged(c.GetWifiDevices(c.GetDevices()))
	if _, ok := ifaces["wlan0"]; ok {
		c.SetIfaceManaged("wlan0", false, c.GetWifiDevices(c.GetDevices()))
	}
}

// manage sets wlan0 to not managed by network manager
func manage(c *netman.Client) {
	c.SetIfaceManaged("wlan0", true, c.GetWifiDevices(c.GetDevices()))
}

// RefreshHandler handles ssids refreshment
func RefreshHandler(w http.ResponseWriter, r *http.Request) {

	c := netman.DefaultClient()
	cw := wifiap.DefaultClient()

	unmanage(c)

	apUp, err := cw.Enabled()
	if err != nil {
		fmt.Println(Sprintf("An error happened while requesting current AP status: %v\n", err))
		return
	}

	if apUp {
		err := cw.Disable()
		if err != nil {
			fmt.Println(Sprintf("An error happened while bringing AP down: %v\n", err))
			return
		}
	}

	for found := scanSsids(utils.SsidsFile, c); !found; found = scanSsids(utils.SsidsFile, c) {
		time.Sleep(5 * time.Second)
	}

	unmanage(c)

	err = cw.Enable()
	if err != nil {
		fmt.Println(Sprintf("An error happened while bringing AP up: %v\n", err))
		return
	}

	// Try to update UI, though it won't probably be possible as far as it is needed to bring down/up AP
	// in a step before.
	ManagementHandler(w, r)
}

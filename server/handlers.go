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
	"log"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"launchpad.net/wifi-connect/utils"
)

const (
	managementTemplatePath  = "/templates/management.html"
	connectingTemplatePath  = "/templates/connecting.html"
	operationalTemplatePath = "/templates/operational.html"
)

// ResourcesPath absolute path to web static resources
var ResourcesPath = filepath.Join(os.Getenv("SNAP"), "static")

var cw interface{}

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

type noData struct {
}

func execTemplate(w http.ResponseWriter, templatePath string, data Data) {
	templateAbsPath := filepath.Join(ResourcesPath, templatePath)
	t, err := template.ParseFiles(templateAbsPath)
	if err != nil {
		log.Printf("Error loading the template at %v: %v", templatePath, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		log.Printf("Error executing the template at %v: %v", templatePath, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// ManagementHandler lists the current available SSIDs
func ManagementHandler(w http.ResponseWriter, r *http.Request) {
	// daemon stores current available ssids in a file
	ssids, err := utils.ReadSsidsFile()
	if err != nil {
		log.Printf("Error reading SSIDs file: %v", err)
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

	pwd := ""
	pwds := r.Form["pwd"]
	if len(pwds) > 0 {
		pwd = pwds[0]
	}

	ssids := r.Form["ssid"]
	if len(ssids) == 0 {
		log.Print("SSID not available")
		return
	}
	ssid := ssids[0]

	data := ConnectingData{ssid}
	execTemplate(w, connectingTemplatePath, data)

	go func() {
		log.Printf("Connecting to %v", ssid)

		err := wifiapClient.Disable()
		if err != nil {
			log.Printf("Error disabling AP: %v", err)
			return
		}

		//connect
		netmanClient.SetIfaceManaged("wlan0", true, netmanClient.GetWifiDevices(netmanClient.GetDevices()))
		_, ap2device, ssid2ap := netmanClient.Ssids()

		err = netmanClient.ConnectAp(ssid, pwd, ap2device, ssid2ap)
		//TODO signal user in portal on failure to connect
		if err != nil {
			log.Printf("Failed connecting to %v.", ssid)
			return
		}

		//remove flag file so that daemon starts checking state
		//and takes control again
		waitPath := os.Getenv("SNAP_COMMON") + "/startingApConnect"
		utils.RemoveFlagFile(waitPath)
	}()
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
		log.Printf("HashItHandler: error hashing: %v", errH)
		return
	}
	res := &hashResponse{}
	res.HashMatch = hashed
	res.Err = "no error"
	b, err := json.Marshal(res)
	if err != nil {
		log.Printf("HashItHandler: error marshaling json")
		return
	}
	w.Write(b)
}

// DisconnectHandler allows user to disconnect from external AP
func DisconnectHandler(w http.ResponseWriter, r *http.Request) {
	netmanClient.DisconnectWifi(netmanClient.GetWifiDevices(netmanClient.GetDevices()))
}

// RefreshHandler handles ssids refreshment
func RefreshHandler(w http.ResponseWriter, r *http.Request) {

	// show same page. After refresh operation, management page should show a refresh alert
	ManagementHandler(w, r)

	go func() {
		if err := netmanClient.Unmanage(); err != nil {
			fmt.Println(err)
			return
		}

		apUp, err := wifiapClient.Enabled()
		if err != nil {
			fmt.Println(Sprintf("An error happened while requesting current AP status: %v\n", err))
			return
		}

		if apUp {
			err := wifiapClient.Disable()
			if err != nil {
				fmt.Println(Sprintf("An error happened while bringing AP down: %v\n", err))
				return
			}
		}

		for found := netmanClient.ScanAndWriteSsidsToFile(utils.SsidsFile); !found; found = netmanClient.ScanAndWriteSsidsToFile(utils.SsidsFile) {
			time.Sleep(5 * time.Second)
		}

		if err := netmanClient.Unmanage(); err != nil {
			fmt.Println(err)
			return
		}

		err = wifiapClient.Enable()
		if err != nil {
			fmt.Println(Sprintf("An error happened while bringing AP up: %v\n", err))
			return
		}

		fmt.Println("== wifi-connect/RefreshHandler: starting wifi-ap")
	}()
}

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

	"strconv"

	"launchpad.net/wifi-connect/utils"
)

const (
	managementTemplatePath  = "/templates/management.html"
	connectingTemplatePath  = "/templates/connecting.html"
	operationalTemplatePath = "/templates/operational.html"
	firstConfigTemplatePath = "/templates/config.html"
)

// ResourcesPath absolute path to web static resources
var ResourcesPath = filepath.Join(os.Getenv("SNAP"), "static")

// first time management portal is accessed this file is created.
var firstConfigFlagFile = filepath.Join(os.Getenv("SNAP_COMMON"), ".first_config")

var cw interface{}

// Data interface representing any data included in a template
type Data interface{}

// ManagementData dynamic data to fulfill the management page.
// It can contain SSIDs list or snap configuration
// NOTE: page is a workaround to render proper grid depending on its value (ssids or config).
// In case authentication mechanism is modified for not to be so intrusive, we could have
// different pages for this
type ManagementData struct {
	Ssids  []string
	Config *utils.Config
	Page   string
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
	config, err := utils.ReadConfig()
	if err != nil {
		log.Printf("Error reading configuration: %v\n", err)
		return
	}

	data := ManagementData{}

	if !config.Portal.NoResetCredentials && utils.MustSetConfig() {
		data.Page = "config"
		data.Config = config
	} else {
		// daemon stores current available ssids in a file
		ssids, err := utils.ReadSsidsFile()
		if err != nil {
			fmt.Printf("== wifi-connect/handler: Error reading SSIDs file: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data.Page = "ssids"
		data.Ssids = ssids
	}

	// parse template
	execTemplate(w, managementTemplatePath, data)
}

// SaveConfigHandler saves config received as form post parameters
func SaveConfigHandler(w http.ResponseWriter, r *http.Request) {

	//TODO TRACE
	log.Println("SAVECONFIGHANDLER")

	// read previous config
	config, err := utils.ReadConfig()
	if err != nil {
		log.Printf("Error reading previous stored config: %v", err)
		return
	}

	r.ParseForm()

	config.Wifi.Ssid = utils.ParseFormParamSingleValue(r.Form, "Ssid")
	config.Wifi.Passphrase = utils.ParseFormParamSingleValue(r.Form, "Passphrase")
	config.Wifi.Interface = utils.ParseFormParamSingleValue(r.Form, "Interface")
	config.Wifi.CountryCode = utils.ParseFormParamSingleValue(r.Form, "CountryCode")
	config.Wifi.Channel, err = strconv.Atoi(utils.ParseFormParamSingleValue(r.Form, "Channel"))
	if err != nil {
		log.Printf("Error parsing channel form value: %v", err)
		return
	}
	config.Wifi.OperationMode = utils.ParseFormParamSingleValue(r.Form, "OperationalMode")
	config.Portal.Password = utils.ParseFormParamSingleValue(r.Form, "PortalPassword")
	showOperational, err := strconv.ParseBool(utils.ParseFormParamSingleValue(r.Form, "ShowOperational"))
	if err != nil {
		log.Printf("Error parsing show operational form value: %v", err)
	}
	// as form received value is 'show_operational', config stored value is the opposite
	config.Portal.NoOperational = !showOperational

	err = utils.WriteConfig(config)
	if err != nil {
		log.Printf("Error saving config: %v", err)
		return
	}

	//after saving config, redirect to management portal, showing available ssids
	ssids, err := utils.ReadSsidsFile()
	if err != nil {
		fmt.Printf("== wifi-connect/handler: Error reading SSIDs file: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	execTemplate(w, managementTemplatePath, ManagementData{Ssids: ssids, Page: "ssids"})
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
		log.Print(Sprintf("SSID not available"))
		return
	}
	ssid := ssids[0]

	data := ConnectingData{ssid}
	execTemplate(w, connectingTemplatePath, data)

	go func() {
		log.Print(Sprintf("Connecting to %v\n", ssid))

		err := wifiapClient.Disable()
		if err != nil {
			log.Print(Sprintf("Error disabling AP: %v\n", err))
			return
		}

		//connect
		netmanClient.SetIfaceManaged("wlan0", true, netmanClient.GetWifiDevices(netmanClient.GetDevices()))
		_, ap2device, ssid2ap := netmanClient.Ssids()

		err = netmanClient.ConnectAp(ssid, pwd, ap2device, ssid2ap)
		//TODO signal user in portal on failure to connect
		if err != nil {
			log.Print(Sprintf("Failed connecting to %v.\n", ssid))
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

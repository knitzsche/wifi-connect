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

package daemon

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"launchpad.net/wifi-connect/avahi"
	"launchpad.net/wifi-connect/netman"
	"launchpad.net/wifi-connect/server"
	"launchpad.net/wifi-connect/utils"
	"launchpad.net/wifi-connect/wifiap"
)

// enum to track current system state
const (
	STARTING = 0 + iota
	MANAGING
	OPERATING
	MANUAL
)

var manualFlagPath string
var waitFlagPath string
var previousState = STARTING
var state = STARTING

// PreConfigFile is the path to the file that stores the hash of the portals password
var PreConfigFile = filepath.Join(os.Getenv("SNAP_COMMON"), "pre-config.json")

// PreConfig is the struct representing a configuration
type PreConfig struct {
	Passphrase         string `json:"wifi-ap.passphrase,omitempty"`
	Ssid               string `json:"wifi-ap.ssid,omitempty"`
	Interface          string `json:"wifi-ap.interface,omitempty"`
	Password           string `json:"portal.password,omitempty"`
	Operational        bool   `json:"portal.operational,omitempty"` //whether to show the operational portal
	ResetCredsRequired bool   `json:"portal.reset-creds,omitempty"` //whether user must reset passphrase and password on first use of mgmt portal
}

// Client is the base type for both testing and runtime
type Client struct {
}

// GetClient returns a client for runtime or testing
func GetClient() *Client {
	return &Client{}
}

// used to clase the operational http server
var err error

// GetManualFlagPath returns the current path
func (c *Client) GetManualFlagPath() string {
	return manualFlagPath
}

// SetManualFlagPath sets the current path
func (c *Client) SetManualFlagPath(s string) {
	manualFlagPath = s
}

// GetWaitFlagPath returns the current path
func (c *Client) GetWaitFlagPath() string {
	return waitFlagPath
}

// SetWaitFlagPath sets the current path
func (c *Client) SetWaitFlagPath(s string) {
	waitFlagPath = s
}

// GetPreviousState returns the daemon previous state
func (c *Client) GetPreviousState() int {
	return previousState
}

// SetPreviousState sets daemon previous state
func (c *Client) SetPreviousState(i int) {
	previousState = i
	return
}

// GetState returns the daemon state
func (c *Client) GetState() int {
	return state
}

// SetState sets the daemon state and updates the previous state
func (c *Client) SetState(i int) {
	previousState = state
	state = i
}

// ScanSsids sets wlan0 to be managed and then scans
// for ssids. If found, write the ssids (comma separated)
// to path and return true, else return false.
func (c *Client) ScanSsids(path string, nc *netman.Client) bool {
	c.Manage(nc)
	SSIDs, _, _ := nc.Ssids()
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

// Unmanage sets wlan0 to be Unmanaged by network manager if it
// is managed
func (c *Client) Unmanage(nc *netman.Client) {
	ifaces, _ := nc.WifisManaged(nc.GetWifiDevices(nc.GetDevices()))
	if _, ok := ifaces["wlan0"]; ok {
		nc.SetIfaceManaged("wlan0", false, nc.GetWifiDevices(nc.GetDevices()))
	}
}

// Manage sets wlan0 to not managed by network manager
func (c *Client) Manage(nc *netman.Client) {
	nc.SetIfaceManaged("wlan0", true, nc.GetWifiDevices(nc.GetDevices()))
}

// CheckWaitApConnect returns true if the flag wait file exists
// and false if it does not
func (c *Client) CheckWaitApConnect() bool {
	if _, err := os.Stat(waitFlagPath); os.IsNotExist(err) {
		return false
	}
	return true
}

// ManualMode enables the daemon to loop without action if in manual mode
// It returns true if the manual mode flag wait file exists
// and false if it does not. If it does not exist and the mode is MANUAL, the
// state is set to STARTING. If it does exist and the mode is not MANUAL, state
// is set to MANUAL
func (c *Client) ManualMode() bool {
	if _, err := os.Stat(manualFlagPath); os.IsNotExist(err) {
		if state == MANUAL {
			c.SetState(STARTING)
			fmt.Println("== wifi-connect: entering STARTING mode")
		}
		return false
	}
	if state != MANUAL {
		c.SetState(MANUAL)
		fmt.Println("== wifi-connect: entering MANUAL mode")
	}
	return true
}

// IsApUpWithoutSSIDs corrects an possible but unlikely case.
// if wifiap is UP and there are no known SSIDs, bring it down so on next
// loop iter we start again and can get SSIDs. returns true when ip is
// UP and has no ssids
func (c *Client) IsApUpWithoutSSIDs(cw *wifiap.Client) bool {
	wifiUp, _ := cw.Enabled()
	if !wifiUp {
		return false
	}
	ssids, _ := utils.ReadSsidsFile()
	if len(ssids) < 1 {
		fmt.Println("== wifi-connect: wifi-ap is UP but has no SSIDS")
		return true // ap is up with no ssids
	}
	return false
}

// ManagementServerUp starts the management server if it is
// not running
func (c *Client) ManagementServerUp() {
	if server.Current != server.Management && server.State == server.Stopped {
		err = server.StartManagementServer()
		if err != nil {
			fmt.Println("== wifi-connect: Error start Mamagement portal:", err)
		}
		// init mDNS
		avahi.InitMDNS()
	}
}

// ManagementServerDown stops the management server if it is running
// also remove the wait flag file, thus resetting proper State
func (c *Client) ManagementServerDown() {
	if server.Current == server.Management && (server.State == server.Running || server.State == server.Starting) {
		err = server.ShutdownManagementServer()
		if err != nil {
			fmt.Println("== wifi-connect: Error stopping the Management portal:", err)
		}
		//remove flag fie so daemon resumes normal control
		utils.RemoveFlagFile(os.Getenv("SNAP_COMMON") + "/startingApConnect")
	}
}

// OperationalServerUp starts the operational server if it is
// not running
func (c *Client) OperationalServerUp() {
	if server.Current != server.Operational && server.State == server.Stopped {
		err = server.StartOperationalServer()
		if err != nil {
			fmt.Println("== wifi-connect: Error starting the Operational portal:", err)
		}
		// init mDNS
		avahi.InitMDNS()
	}
}

// OperationalServerDown stops the operational server if it is running
func (c *Client) OperationalServerDown() {
	if server.Current == server.Operational && (server.State == server.Running || server.State == server.Starting) {
		err = server.ShutdownOperationalServer()
		if err != nil {
			fmt.Println("== wifi-connect: Error stopping Operational portal:", err)
		}
	}
}

// SetDefaults checks if there is a configuration file, and if so it applies the configuration,
// returning true, nil on success, true, error on failure. If there is no configuration file,
// false, error is returned
func (c *Client) SetDefaults() (bool, error) {
	fmt.Println("== wifi-connect/SetDefaults running")
	_, err := os.Open(PreConfigFile)
	if err != nil {
		return false, err
	}
	content, errR := ioutil.ReadFile(ConfigFile)
	if errR != nil {
		return true, err
	}
	config := &PreConfig{}
	errJ := json.Unmarshal(content, config)
	if errJ != nil {
		return true, errJ
	}
	if len(config.Passphrase) > 0 {
		fmt.Println("== wifi-connect/SetDefaults wifi-ap passphrase being set")
		c := wifiap.DefaultClient()
		err := c.SetPassphrase(config.Passphrase)
		if err != nil {
			return true, err
		}
	}

	if len(config.Ssid) > 0 {
		fmt.Println("== wifi-connect/SetDefaults wifi-ap SSID being set to", config.Ssid)
		c := wifiap.DefaultClient()
		err := c.SetSsid(config.Ssid)
		if err != nil {
			return true, err
		}
	}
	if len(config.Interface) > 0 {
		fmt.Println("== wifi-connect/SetDefaults wifi-ap interface being set to", config.Interface)
		//TODO implementation
	}
	if len(config.Password) > 0 {
		fmt.Println("== wifi-connect/SetDefaults portal password being set")
		utils.HashIt(config.Password)
	}
	if !config.Opertional > 0 {
		fmt.Println("== wifi-connect/SetDefaults: don't show opertionational portal")
		// TODO
	}
	if !config.ResetCredsRequired > 0 {
		fmt.Println("== wifi-connect/SetDefaults: don't require creds are reset")
		// TODO
	}

	fmt.Println("== wifi-connect/SetDefaults done")
	return true, nil
}

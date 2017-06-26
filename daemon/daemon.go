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

	"launchpad.net/wifi-connect/avahi"
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
	PreConfigFile bool   `json:"config.file,omitempty"`
	Passphrase    string `json:"wifi.security-passphrase,omitempty"`
	Ssid          string `json:"wifi.ssid,omitempty"`
	Interface     string `json:"wifi.interface,omitempty"`
	Password      string `json:"portal.password,omitempty"`
	NoOperational bool   `json:"portal.no-operational,omitempty"` //whether to show the operational portal
	NoResetCreds  bool   `json:"portal.no-reset-creds,omitempty"` //whether user must reset passphrase and password on first use of mgmt portal
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

// NewConfig makes a new preconfiguration - use ONLY for testing
func (c *Client) NewConfig() *PreConfig {
	return &PreConfig{}
}

// SetDefaults creates the run time configuration based on wifi-ap and the pre-config.json
// configuration file, if any. The configuration is returned with an error. PreConfig.PreConfigfile
// indicates whether a pre-config file exists.
func (c *Client) SetDefaults(cw wifiap.Operations) (*PreConfig, error) {
	config := &PreConfig{PreConfigFile: true}
	content, err := ioutil.ReadFile(PreConfigFile)
	if err != nil {
		config.PreConfigFile = false
	}
	err = json.Unmarshal(content, config)
	if err != nil {
		return config, err
	}
	ap, errShow := cw.Show()
	if errShow != nil {
		fmt.Println("== wifi-connect/daemon/SetDefaults: wifi-ap.Show err:", errShow)
	}
	if ap["wifi.security-passphrase"] != config.Passphrase {
		if len(config.Passphrase) > 0 {
			err = cw.SetPassphrase(config.Passphrase)
			fmt.Println("== wifi-connect/SetDefaults wifi-ap passphrase being set")
			if err != nil {
				fmt.Println("== wifi-connect/daemon/SetDefaults: passphrase err:", err)
				return config, err
			}
		}
	}
	if len(config.Password) > 0 {
		fmt.Println("== wifi-connect/SetDefaults portal password being set")
		_, err = utils.HashIt(config.Password)
		if err != nil {
			fmt.Println("== wifi-connect/daemon/SetDefaults: password err:", err)
			return config, err
		}
	}
	if config.NoOperational {
		fmt.Println("== wifi-connect/SetDefaults: operational portal is now disabled")
	}
	if config.NoResetCreds {
		fmt.Println("== wifi-connect/SetDefaults: reset creds requirement is now disabled")
	}
	return config, nil
}

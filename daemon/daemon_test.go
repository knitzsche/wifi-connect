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
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"launchpad.net/wifi-connect/utils"
)

func TestManualFlagPath(t *testing.T) {
	mfpInit := "mfp"
	client := GetClient()
	client.SetManualFlagPath(mfpInit)
	mfp := client.GetManualFlagPath()
	if mfp != mfpInit {
		t.Errorf("ManualFlag path should be %s but is %s\n", mfpInit, mfp)
	}
}

func TestWaitFlagPath(t *testing.T) {
	wfpInit := "wfp"
	client := GetClient()
	client.SetWaitFlagPath(wfpInit)
	wfp := client.GetWaitFlagPath()
	if wfp != wfpInit {
		t.Errorf("WaitFlag path should be %s but is %s", wfp, wfpInit)
	}
}

func TestState(t *testing.T) {
	client := GetClient()
	client.SetState(MANAGING)
	client.SetPreviousState(STARTING)
	ps := client.GetPreviousState()
	if ps != STARTING {
		t.Errorf("Previous state should be %d but is %d", STARTING, ps)
	}
	s := client.GetState()
	if s != MANAGING {
		t.Errorf("State should be %d but is %d", MANAGING, s)
	}
}

func TestCheckWaitApConnect(t *testing.T) {
	client := GetClient()
	wfp := "thispathnevershouldexist"
	client.SetWaitFlagPath(wfp)
	if client.CheckWaitApConnect() {
		t.Errorf("CheckWaitApConnect returns true but should return false")
	}
	wfp = "../static/tests/waitFile"
	client.SetWaitFlagPath(wfp)
	if !client.CheckWaitApConnect() {
		t.Errorf("CheckWaitApConnect returns false but should return true")
	}
}

func TestManualMode(t *testing.T) {
	mfp := "thisfileshouldneverexist"
	client := GetClient()
	client.SetManualFlagPath(mfp)
	client.SetState(MANUAL)
	if client.ManualMode() {
		t.Errorf("ManualMode returns true but should return false")
	}
	if client.GetState() != STARTING {
		t.Errorf("ManualMode should set state to STARTING when not in manual mode but does not")
	}
	mfp = "../static/tests/manualMode"
	client.SetManualFlagPath(mfp)
	client.SetState(STARTING)
	if !client.ManualMode() {
		t.Errorf("ManualMode returns false but should return true")
	}
	if client.GetState() != MANUAL {
		t.Errorf("ManualMode should set state to MANUAL when in manual mode but does not")
	}
}

type mockWifiap struct{}

func (mock *mockWifiap) Do(req *http.Request) (*http.Response, error) {
	fmt.Println("==== MY do called")
	url := req.URL.String()
	if url != "http://unix/v1/configuration" {
		return nil, fmt.Errorf("Not valid request URL: %v", url)
	}

	if req.Method != "GET" {
		return nil, fmt.Errorf("Method is not valid. Expected GET, got %v", req.Method)
	}

	rawBody := `{"result":{
		"debug":false, 
		"dhcp.lease-time": "12h", 
		"dhcp.range-start": "10.0.60.2", 
		"dhcp.range-stop": "10.0.60.199", 
		"disabled": true, 
		"share.disabled": false, 
		"share-network-interface": "tun0", 
		"wifi-address": "10.0.60.1", 
		"wifi.channel": "6", 
		"wifi.hostapd-driver": "nl80211", 
		"wifi.interface": "wlan0", 
		"wifi.interface-mode": "direct", 
		"wifi.netmask": "255.255.255.0", 
		"wifi.operation-mode": "g", 
		"wifi.security": "
		"wifi.security-passphrase": "passphrase123", 
		"wifi.ssid": "AP"},"status":"OK","status-code":200,""sync"}`

	response := http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Body:       ioutil.NopCloser(strings.NewReader(rawBody)),
	}

	return &response, nil
}

func (mock *mockWifiap) Show() (map[string]interface{}, error) {
	wifiAp := make(map[string]interface{})
	wifiAp["wifi.security-passphrase"] = "randompassphrase"
	return wifiAp, nil
}

func (mock *mockWifiap) Enabled() (bool, error) {
	return true, nil
}

func (mock *mockWifiap) Enable() error {
	return nil
}
func (mock *mockWifiap) Disable() error {
	return nil
}
func (mock *mockWifiap) SetSsid(s string) error {
	return nil
}

func (mock *mockWifiap) SetPassphrase(p string) error {
	return nil
}

func TestLoadPreConfig(t *Testing) {
	client := GetClient()
	PreConfigFile = "../static/tests/pre-config0.json"
	config, err := LoadPreConfig()
	if err != nil {
		t.Errorf("Unexpected error using  LoadPreConfig:", err)
	}
	if config.Passphrase != "abcdefghijklmnop" {
		t.Errorf("Passphrase of %s expected but got %s:", "abcdefghijklmnop", config.Passphrase)
	}
	if !config.no - operational {
		t.Errorf("portal.no-operational was set to true but the loaded config is %t", config.NoOperational)
	}
	if !config.no - reset - creds {
		t.Errorf("portal.no-reset-creds was set to true but the loaded config is %t", config.NoResetCreds)
	}

}

func TestSetDefaults(t *testing.T) {
	client := GetClient()
	PreConfigFile = "../static/tests/pre-config0.json"
	hfp := "/tmp/hash"
	if _, err := os.Stat(hfp); err == nil {
		err = os.Remove(hfp)
		if err != nil {
			t.Errorf("Could not remove previous file version")
		}
	}
	utils.SetHashFile(hfp)
	config, _ := client.SetDefaults(&mockWifiap{})
	expectedPassphrase := "abcdefghijklmnop"
	expectedPassword := "qwerzxcv"
	if config.Passphrase != expectedPassphrase {
		t.Errorf("SetDefaults: Preconfig passphrase should be %s but is %s", expectedPassphrase, config.Passphrase)
	}
	if os.IsNotExist(err) {
		t.Errorf("SetDefaults should have created %s but did not", hfp)
	}
	res, _ := utils.MatchingHash(expectedPassword)
	if !res {
		t.Errorf("SetDefaults: Preconfig password hash did not match actual")
	}
	if !config.NoOperational {
		t.Errorf("SetDefaults: Preconfig portal.no-operational should be true (set) but is %t", config.NoOperational)
	}
	if !config.NoResetCreds {
		t.Errorf("SetDefaults: Preconfig portal.no-reset-creds should be true (set) but is %t", config.NoResetCreds)
	}

	if _, err := os.Stat(hfp); err == nil {
		err = os.Remove(hfp)
		if err != nil {
			t.Errorf("Could not remove previous file version")
		}
	}
	PreConfigFile = "../static/tests/pre-config1.json"
	config, _ = client.SetDefaults(&mockWifiap{})
	if len(config.Passphrase) > 0 {
		t.Errorf("SetDefaults: Preconfig passphrase was not set but is %s", config.Passphrase)
	}
	res2, _ := utils.MatchingHash(expectedPassword)
	if res2 {
		t.Errorf("SetDefaults: Preconfig password was not set, but the hash matched")
	}
	if config.NoOperational {
		t.Errorf("SetDefaults: Preconfig portal.no-operational should be false (unset) but is %t", config.NoOperational)
	}
	if config.NoResetCreds {
		t.Errorf("SetDefaults: Preconfig portal.no-reset-creds should be false (unnset) but is %t", config.NoResetCreds)
	}
}

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
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"launchpad.net/wifi-connect/netman"
	"launchpad.net/wifi-connect/utils"
)

type wifiapClientMock struct{}

func (c *wifiapClientMock) Show() (map[string]interface{}, error) {
	return nil, nil
}

func (c *wifiapClientMock) Enable() error {
	return nil
}

func (c *wifiapClientMock) Disable() error {
	return nil
}

func (c *wifiapClientMock) Enabled() (bool, error) {
	return true, nil
}

func (c *wifiapClientMock) SetSsid(string) error {
	return nil
}

func (c *wifiapClientMock) SetPassphrase(string) error {
	return nil
}

func (c *wifiapClientMock) Set(map[string]interface{}) error {
	return nil
}

type netmanClientMock struct{}

func (c *netmanClientMock) GetDevices() []string {
	return []string{"/d/1"}
}

func (c *netmanClientMock) GetWifiDevices(devices []string) []string {
	return []string{"/d/1"}
}

func (c *netmanClientMock) GetAccessPoints(devices []string, ap2device map[string]string) []string {
	return []string{"/ap/1"}
}

func (c *netmanClientMock) ConnectAp(ssid string, p string, ap2device map[string]string, ssid2ap map[string]string) error {
	return nil
}

func (c *netmanClientMock) Ssids() ([]netman.SSID, map[string]string, map[string]string) {
	myssid := netman.SSID{Ssid: "myssid", ApPath: "/ap/1"}
	return []netman.SSID{myssid}, map[string]string{"/ap/1": "/d/1"}, map[string]string{"myssid": "/ap/1"}
}

func (c *netmanClientMock) Connected(devices []string) bool {
	return false
}

func (c *netmanClientMock) ConnectedWifi(wifiDevices []string) bool {
	return false
}

func (c *netmanClientMock) DisconnectWifi(wifiDevices []string) int {
	return 0
}

func (c *netmanClientMock) SetIfaceManaged(iface string, state bool, devices []string) string {
	return "wlan0"
}

func (c *netmanClientMock) WifisManaged(wifiDevices []string) (map[string]string, error) {
	return map[string]string{"wlan0": "/d/1"}, nil
}
func (c *netmanClientMock) Unmanage() error {
	return nil
}
func (c *netmanClientMock) Manage() error {
	return nil
}

func (c *netmanClientMock) ScanAndWriteSsidsToFile(filepath string) bool {
	return true
}

func TestManagementHandler(t *testing.T) {

	ResourcesPath = "../static"
	SsidsFile := "../static/tests/ssids"
	utils.SetSsidsFile(SsidsFile)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	http.HandlerFunc(ManagementHandler).ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got: %d", http.StatusOK, w.Code)
	}

	if !strings.Contains(w.Header().Get("Content-Type"), "text/html") {
		t.Error("Response content type is not expected text/html")
	}
}

func TestConnectHandler(t *testing.T) {

	wifiapClient = &wifiapClientMock{}
	netmanClient = &netmanClientMock{}

	ResourcesPath = "../static"

	form := url.Values{}
	form.Add("ssid", "myssid")
	form.Add("pwd", "mypassphrase")

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/connect", bytes.NewBufferString(form.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(form.Encode())))
	http.HandlerFunc(ConnectHandler).ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got: %d", http.StatusOK, w.Code)
	}

	if !strings.Contains(w.Header().Get("Content-Type"), "text/html") {
		t.Error("Response content type is not expected text/html")
	}
}

func TestDisconnectHandler(t *testing.T) {

	netmanClient = &netmanClientMock{}

	ResourcesPath = "../static"

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/disconnect", nil)
	http.HandlerFunc(DisconnectHandler).ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got: %d", http.StatusOK, w.Code)
	}
}

func TestRefreshHandler(t *testing.T) {

	wifiapClient = &wifiapClientMock{}
	netmanClient = &netmanClientMock{}

	ResourcesPath = "../static"
	SsidsFile := "../static/tests/ssids"
	utils.SetSsidsFile(SsidsFile)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/refresh", nil)
	http.HandlerFunc(RefreshHandler).ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got: %d", http.StatusOK, w.Code)
	}

	if !strings.Contains(w.Header().Get("Content-Type"), "text/html") {
		t.Error("Response content type is not expected text/html")
	}
}

func TestOperationalHandler(t *testing.T) {
	ResourcesPath = "../static"

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	http.HandlerFunc(OperationalHandler).ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got: %d", http.StatusOK, w.Code)
	}

	if !strings.Contains(w.Header().Get("Content-Type"), "text/html") {
		t.Error("Response content type is not expected text/html")
	}
}

func TestInvalidTemplateHandler(t *testing.T) {

	ResourcesPath = "/invalidpath"

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	http.HandlerFunc(ManagementHandler).ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got: %d", http.StatusInternalServerError, w.Code)
	}
}

func TestReadSsidsFile(t *testing.T) {

	SsidsFile := "../static/tests/ssids"
	utils.SetSsidsFile(SsidsFile)
	ssids, err := utils.ReadSsidsFile()
	if err != nil {
		t.Errorf("Unexpected error reading ssids file: %v", err)
	}

	if len(ssids) != 4 {
		t.Error("Expected 4 elements in csv record")
	}

	set := make(map[string]bool)
	for _, v := range ssids {
		set[v] = true
	}

	if !set["mynetwork"] {
		t.Error("mynetwork value not found")
	}
	if !set["yournetwork"] {
		t.Error("yournetwork value not found")
	}
	if !set["hernetwork"] {
		t.Error("hernetwork value not found")
	}
	if !set["hisnetwork"] {
		t.Error("hisnetwork value not found")
	}
}

func TestReadSsidsFileWithOnlyOne(t *testing.T) {

	SsidsFile := "../static/tests/ssids_onlyonessid"
	utils.SetSsidsFile(SsidsFile)
	ssids, err := utils.ReadSsidsFile()
	if err != nil {
		t.Errorf("Unexpected error reading ssids file: %v", err)
	}

	if len(ssids) != 1 {
		t.Error("Expected 1 elements in csv record")
	}

	set := make(map[string]bool)
	for _, v := range ssids {
		set[v] = true
	}

	if !set["mynetwork"] {
		t.Error("mynetwork value not found")
	}
}

func TestReadEmptySsidsFile(t *testing.T) {

	SsidsFile := "../static/tests/ssids_empty"
	utils.SetSsidsFile(SsidsFile)
	ssids, err := utils.ReadSsidsFile()
	if err != nil {
		t.Errorf("Unexpected error reading ssids file: %v", err)
	}

	if len(ssids) != 0 {
		t.Error("Expected 0 elements in csv record")
	}
}

// -*- Mode: Go; indent-tabs-mode: t -*-

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
	"net/http"

	"launchpad.net/wifi-connect/netman"
	"launchpad.net/wifi-connect/wifiap"
)

// Operations interface defining operations implemented by wifiap client
type wifiapOperations interface {
	Show() (map[string]interface{}, error)
	Enabled() (bool, error)
	Enable() error
	Disable() error
	SetSsid(string) error
	SetPassphrase(string) error
}

type netmanOperations interface {
	GetDevices() []string
	GetWifiDevices(devices []string) []string
	GetAccessPoints(devices []string, ap2device map[string]string) []string
	ConnectAp(ssid string, p string, ap2device map[string]string, ssid2ap map[string]string) error
	Ssids() ([]netman.SSID, map[string]string, map[string]string)
	Connected(devices []string) bool
	ConnectedWifi(wifiDevices []string) bool
	DisconnectWifi(wifiDevices []string) int
	SetIfaceManaged(iface string, state bool, devices []string) string
	WifisManaged(wifiDevices []string) (map[string]string, error)
	Unmanage() error
	Manage() error
	ScanAndWriteSsidsToFile(filepath string) bool
}

var wifiapClient wifiapOperations
var netmanClient netmanOperations

// Middleware to pre-process web service requests
func Middleware(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if wifiapClient == nil {
			wifiapClient = wifiap.DefaultClient()
		}

		if netmanClient == nil {
			netmanClient = netman.DefaultClient()
		}

		inner.ServeHTTP(w, r)
	})
}

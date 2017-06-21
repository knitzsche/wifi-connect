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

var wifiapClient wifiap.Operations
var netmanClient netman.Operations

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

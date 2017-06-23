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

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"

	"launchpad.net/wifi-connect/daemon"
)

type Client struct {
	getter Getter
}

type Get struct{}

type Getter interface {
	SnapGet(string) (string, error)
}

func GetClient() *Client {
	return &Client{getter: &Get{}}
}

func GetTestClient(g Getter) *Client {
	return &Client{getter: g}
}

func (g *Get) SnapGet(key string) (string, error) {
	out, err := exec.Command("snapctl", "get", key).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil

}

func (c *Client) snapGetStr(key string, target *string) {
	val, err := c.getter.SnapGet(key)
	if err != nil {
		return
	}
	if len(val) > 0 {
		*target = val
		return
	} else {
		log.Printf("== wifi-connect/configure error: key %s exists but has zero length", key)
	}
}

func (c *Client) snapGetBool(key string, target *bool) {
	val, err := c.getter.SnapGet(key)
	if err != nil {
		return
	}
	if len(val) > 0 {
		if val == "true" {
			*target = true
			return
		} else {
			*target = false
			return
		}
	} else {
		log.Printf("== wifi-connect/configure error: key %s exists but has zero length", key)
	}

}

func main() {
	client := GetClient()
	preConfig := &daemon.PreConfig{}
	client.snapGetStr("wifi.security-passphrase", &preConfig.Passphrase)
	client.snapGetStr("portal.password", &preConfig.Password)
	client.snapGetBool("portal.no-operational", &preConfig.NoOperational)
	client.snapGetBool("portal.no-reset-creds", &preConfig.NoResetCreds)

	b, errJM := json.Marshal(preConfig)
	if errJM == nil {
		errWJ := ioutil.WriteFile(daemon.PreConfigFile, b, 0644)
		if errWJ != nil {
			log.Print("== wifi-connect/configure error:", errWJ)
			return
		}
	}
}

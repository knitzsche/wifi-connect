// -*- Mode: Go; indent-tabs-mode: t -*-

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"

	"launchpad.net/wifi-connect/daemon"
)

func snapGet(key string) (string, error) {
	out, err := exec.Command("snapctl", "get", key).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil

}

func main() {
	preConfig := &daemon.PreConfig{}
	var val string
	var err error
	val, err = snapGet("wifi.security-passphrase")
	if err != nil {
		log.Print("== wifi-connect/configure error", err)
	}
	if len(val) > 0 {
		preConfig.Passphrase = val
	}

	val, err = snapGet("portal.password")
	if err != nil {
		log.Print("== wifi-connect/configure error", err)
	}
	if len(val) > 0 {
		preConfig.Password = val
	}

	val, err = snapGet("portal.no-operational")
	if err != nil {
		log.Print("== wifi-connect/configure error", err)
	}
	preConfig.NoOperational = false // default
	if val == "true" {
		preConfig.NoOperational = true
	}

	val, err = snapGet("portal.no-reset-creds")
	if err != nil {
		log.Print("== wifi-connect/configure error", err)
	}
	preConfig.NoResetCreds = false // default
	if val == "true" {
		preConfig.NoResetCreds = true
	}

	b, errJM := json.Marshal(preConfig)
	if errJM == nil {
		errWJ := ioutil.WriteFile(daemon.PreConfigFile, b, 0644)
		if errWJ != nil {
			log.Print("== wifi-connect/configure error:", errWJ)
			return
		}
	}
}

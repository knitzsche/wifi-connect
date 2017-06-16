// -*- Mode: Go; indent-tabs-mode: t -*-

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
	val, err = snapGet("passphrase")
	if err != nil {
		log.Print("== wifi-connect/configure error", err)
		return
	}
	if len(val) > 0 {
		preConfig.Passphrase = val
	}

	val, err = snapGet("ssid")
	if err != nil {
		log.Print("== wifi-connect/configure error", err)
		return
	}
	if len(val) > 0 {
		preConfig.Ssid = val
	}

	val, err = snapGet("interface")
	if err != nil {
		log.Print("== wifi-connect/configure error", err)
		return
	}
	if len(val) > 0 {
		preConfig.Interface = val
	}

	val, err = snapGet("password")
	if err != nil {
		log.Print("== wifi-connect/configure error", err)
		return
	}
	if len(val) > 0 {
		preConfig.Password = val
	}

	val, err = snapGet("no-operational")
	if err != nil {
		log.Print("== wifi-connect/configure error", err)
		return
	}
	preConfig.NoOperational = false // default
	if val == "true" {
		preConfig.NoOperational = true
	}

	val, err = snapGet("no-reset-creds")
	if err != nil {
		log.Print("== wifi-connect/configure error", err)
		return
	}
	preConfig.NoResetCreds = false // default
	if val == "true" {
		preConfig.NoResetCreds = true
	}

	confFile := filepath.Join(os.Getenv("SNAP_COMMON"), "pre-config.json")

	b, errJM := json.Marshal(preConfig)
	if errJM == nil {
		errWJ := ioutil.WriteFile(confFile, b, 0644)
		if errWJ != nil {
			log.Print("== wifi-connect/configure error:", errWJ)
			return
		}
	}
}

// -*- Mode: Go; indent-tabs-mode: t -*-

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"

	"launchpad.net/wifi-connect/daemon"
)

func snapGet(key string, target interface{}) (interface{}, error) {
	var res interface{}
	var ok bool
	out, err := exec.Command("snapctl", "get", key).Output()
	if err != nil {
		return res, err
	}
	val := strings.TrimSpace(string(out))
	_, ok = target.(string)
	if ok {
		if len(val) > 0 {
			return val, nil
		} else {
			return "", fmt.Errorf("== wifi-connect/configure error: key %s exists but has zero length")
		}
	}

	_, ok = target.(bool)
	if ok {
		if len(val) > 0 {
			if val == "true" {
				return true, nil
			} else {
				return false, nil
			}
		} else {
			return res, fmt.Errorf("== wifi-connect/configure error: key %s exists but has zero length")
		}
	}
	return res, nil
}

func main() {
	preConfig := &daemon.PreConfig{}
	var res interface{}
	var err error
	res, err = snapGet("wifi.security-passphrase", preConfig.Passphrase)
	if err != nil {
		log.Print(err)
	} else {
		preConfig.Passphrase = res.(string)
	}

	res, err = snapGet("portal.password", preConfig.Password)
	if err != nil {
		log.Print(err)
	} else {
		preConfig.Password = res.(string)
	}

	res, err = snapGet("portal.no-operational", preConfig.NoOperational)
	if err != nil {
		log.Print(err)
	} else {
		preConfig.NoOperational = res.(bool)
	}

	res, err = snapGet("portal.no-reset-creds", preConfig.NoResetCreds)
	if err != nil {
		log.Print(err)
	} else {
		preConfig.NoResetCreds = res.(bool)
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

package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"launchpad.net/wifi-connect/wifiap"
)

var configFile = filepath.Join(os.Getenv("SNAP_COMMON"), "config.json")
var mustConfigFlagFile = filepath.Join(os.Getenv("SNAP_COMMON"), ".config_done.flag")

var wifiapClient wifiap.Operations = wifiap.DefaultClient()

// Config this project config got from wifi-ap + custom wifi-connect params
type Config struct {
	Wifi   *WifiConfig
	Portal *PortalConfig
}

// WifiConfig config specific parameters for wifi configuration
type WifiConfig struct {
	Ssid          string `json:"wifi.ssid"`
	Passphrase    string `json:"wifi.security-passphrase"`
	Interface     string `json:"wifi.interface"`
	CountryCode   string `json:"wifi.country-code"`
	Channel       int    `json:"wifi.channel"`
	OperationMode string `json:"wifi.operation-mode"`
}

// PortalConfig config specific parameters for portals configuration
type PortalConfig struct {
	Password           string `json:"portal.password"`
	NoResetCredentials bool   `json:"portal.no-reset-creds"`
	NoOperational      bool   `json:"portal.no-operational"`
}

func (c *Config) String() string {
	s := []string{
		strings.Join([]string{"--Wifi--", c.Wifi.String()}, "\n"),
		strings.Join([]string{"--Portal--", c.Portal.String()}, "\n"),
	}
	return fmt.Sprintf(strings.Join(s, "\n"))
}

func (c *PortalConfig) String() string {
	s := []string{
		strings.Join([]string{"Password: ", c.Password}, " "),
		strings.Join([]string{"NoResetCredentials:", strconv.FormatBool(c.NoResetCredentials)}, " "),
		strings.Join([]string{"NoOperational:", strconv.FormatBool(c.NoOperational)}, " "),
	}
	return fmt.Sprintf(strings.Join(s, "\n"))
}

func (c *WifiConfig) String() string {
	s := []string{
		strings.Join([]string{"Ssid: ", c.Ssid}, " "),
		strings.Join([]string{"Passphrase:", c.Passphrase}, " "),
		strings.Join([]string{"Interface:", c.Interface}, " "),
		strings.Join([]string{"CountryCode:", c.CountryCode}, " "),
		strings.Join([]string{"Channel:", strconv.Itoa(c.Channel)}, " "),
		strings.Join([]string{"OperationMode:", c.OperationMode}, " "),
	}
	return fmt.Sprintf(strings.Join(s, "\n"))
}

func defaultPortalConfig() *PortalConfig {
	return &PortalConfig{
		Password:           "",
		NoResetCredentials: false,
		NoOperational:      false,
	}
}

// config currently stored in local json file is completely storable in PortalConfig
// If needed to scale, we could rewrite this method to support a more generic type
func readLocalConfig() (*PortalConfig, error) {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Printf("Not found local config file at %v. Applying default local configuration", configFile)
		return defaultPortalConfig(), nil
	}

	fileContents, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("Error reading json config file: %v", err)
	}

	portalConfig := defaultPortalConfig()
	err = json.Unmarshal(fileContents, portalConfig)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling json config file contents: %v", err)
	}

	return portalConfig, nil
}

func writeLocalConfig(p *PortalConfig) error {
	bytes, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("Could not marshal local config to raw data: %v", err)
	}

	err = ioutil.WriteFile(configFile, bytes, 0644)
	if err != nil {
		return fmt.Errorf("Could not write local config to file: %v", err)
	}

	// write flag file for not asking more times for configuring snap before first use
	if MustSetConfig() {
		err = WriteFlagFile(mustConfigFlagFile)
		if err != nil {
			return fmt.Errorf("Error writing flag file after configuring for a first time")
		}
	}

	return nil
}

func readRemoteParam(m map[string]interface{}, key string, defaultValue interface{}) interface{} {
	val, ok := m[key]
	if !ok {
		val = defaultValue
		log.Printf("Warning: %v key was not found in remote config", key)
	}

	return val
}

func readRemoteConfig() (*WifiConfig, error) {
	settings, err := wifiapClient.Show()
	if err != nil {
		return nil, fmt.Errorf("Error reading wifi-ap remote configuration: %v", err)
	}

	// NOTE: Preprocessing for the case of wifi.channel.
	// In the case of the channel, it is returned as string from rest api, but we have to convert it
	// as it is handled as int internally.
	// In case wifi-ap provides this in future as int, this could be replaced by
	//
	// readRemoteParam(settings, "wifi.channel", 0).(int)
	channel, err := strconv.Atoi(readRemoteParam(settings, "wifi.channel", "0").(string))
	if err != nil {
		return nil, fmt.Errorf("Could not parse wifi.channel parameter: %v", err)
	}

	return &WifiConfig{
		Ssid:          readRemoteParam(settings, "wifi.ssid", "").(string),
		Passphrase:    readRemoteParam(settings, "wifi.security-passphrase", "").(string),
		Interface:     readRemoteParam(settings, "wifi.interface", "").(string),
		CountryCode:   readRemoteParam(settings, "wifi.country-code", "").(string),
		Channel:       channel,
		OperationMode: readRemoteParam(settings, "wifi.operation-mode", "").(string),
	}, nil
}

func writeRemoteConfig(wc *WifiConfig) error {
	params := make(map[string]interface{})
	params["wifi.ssid"] = wc.Ssid
	params["wifi.security-passphrase"] = wc.Passphrase
	params["wifi.interface"] = wc.Interface
	params["wifi.country-code"] = wc.CountryCode
	params["wifi.channel"] = wc.Channel
	params["wifi.operation-mode"] = wc.OperationMode

	err := wifiapClient.Set(params)
	if err != nil {
		return fmt.Errorf("Error writing remote configuration: %v", err)
	}

	return nil
}

// ReadConfig reads all config, remote and local, at the same time
var ReadConfig = func() (*Config, error) {
	wifiConfig, err := readRemoteConfig()
	if err != nil {
		return nil, err
	}

	portalConfig, err := readLocalConfig()
	if err != nil {
		return nil, err
	}

	return &Config{Wifi: wifiConfig, Portal: portalConfig}, nil
}

// WriteConfig writes all remote and local config at the same time
var WriteConfig = func(c *Config) error {
	// read local config and save it as backup. This will be written back
	// in case saving remote config fails
	localConfigBackup, err := readLocalConfig()
	if err != nil {
		return fmt.Errorf("Error reading current config before applying new one: %v", err)
	}

	err = writeLocalConfig(c.Portal)
	if err != nil {
		// if an error happens writing local config there is no need to restore
		// backup, as nothing has been written
		return err
	}

	err = writeRemoteConfig(c.Wifi)
	if err != nil {
		// restore backup
		backupErr := writeLocalConfig(localConfigBackup)
		if backupErr != nil {
			return fmt.Errorf("Could not restore previous local configuration: %v", backupErr)
		}
		// return error after completing backup
		return err
	}

	return nil
}

// MustSetConfig true if one needs to configure snap before continuing
var MustSetConfig = func() bool {
	if _, err := os.Stat(mustConfigFlagFile); os.IsNotExist(err) {
		return true
	}
	return false
}

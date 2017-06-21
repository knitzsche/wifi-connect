package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"launchpad.net/wifi-connect/wifiap"
)

var configFile = filepath.Join(os.Getenv("SNAP_COMMON"), "config.json")

var wifiapClient wifiap.Operations

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
	Regdomain     string `json:"wifi.regdomain"`
	Channel       int    `json:"wifi.channel"`
	OperationMode string `json:"wifi.operation-mode"`
}

// PortalConfig config specific parameters for portals configuration
type PortalConfig struct {
	Password           string `json:"portal.password"`
	NoResetCredentials bool   `json:"portal.no-reset-creds"`
	NoOperational      bool   `json:"portal.no-operational"`
}

// config currently stored in local json file is completely storable in PortalConfig
// If needed to scale, we could rewrite this method to support a more generic type
func readLocalConfig() (*PortalConfig, error) {
	fileContents, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("Error reading json config file: %v", err)
	}

	var portalConfig PortalConfig
	err = json.Unmarshal(fileContents, &portalConfig)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling json config file contents: %v", err)
	}

	return &portalConfig, nil
}

func readRemoteConfig() (*WifiConfig, error) {

	settings, err := wifiapClient.Show()
	if err != nil {
		return nil, fmt.Errorf("Error reading wifi-ap remote configuration: %v", err)
	}

	return &WifiConfig{
		Ssid:          settings["wifi.ssid"].(string),
		Passphrase:    settings["wifi.security-passphrase"].(string),
		Interface:     settings["wifi.interface"].(string),
		Regdomain:     settings["wifi.regdomain"].(string),
		Channel:       settings["wifi.channel"].(int),
		OperationMode: settings["wifi.operational-mode"].(string)}, nil
}

func readConfig() (*Config, error) {
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

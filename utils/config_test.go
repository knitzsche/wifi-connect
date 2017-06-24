package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"gopkg.in/check.v1"
)

const testLocalConfig = `
{
	"portal.password": "the_password",
	"portal.no-reset-creds": true,
	"portal.no-operational": false	
}
`

const testLocalConfigBadEntry = `
{
	"portal.password": "the_password",
	"portal.no-reset-creds": true,
	"bad.parameter": "bad.value",
	"portal.no-operational": false
}
`

const testLocalEmptyConfig = `
{
}
`

var testPortalConfig = &PortalConfig{"the_password", true, false}

var rand uint32
var randmu sync.Mutex

func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

// ####################
// Testing local config
// ####################
func randomName() string {
	randmu.Lock()
	r := rand
	if r == 0 {
		r = uint32(time.Now().UnixNano() + int64(os.Getpid()))
	}
	r = r*1664525 + 1013904223 // constants from Numerical Recipes
	rand = r
	randmu.Unlock()
	return strconv.Itoa(int(1e9 + r%1e9))[1:]
}

func createTempFile(content string) (*os.File, error) {
	contentAsBytes := []byte(content)

	tmpfile, err := ioutil.TempFile("", "config")
	if err != nil {
		return nil, fmt.Errorf("Could not create temp file: %v", err)
	}

	if _, err := tmpfile.Write(contentAsBytes); err != nil {
		return nil, fmt.Errorf("Could not write contents to temp file: %v", err)
	}

	if err := tmpfile.Close(); err != nil {
		return nil, fmt.Errorf("Could not close tempfile properly: %v", err)
	}

	return tmpfile, nil
}

func verifyLocalConfig(c *check.C, cfg *PortalConfig, expectedPwd string, expectedNoResetCredentials bool, expectedNoOperational bool) {
	c.Assert(cfg.Password, check.Equals, expectedPwd)
	c.Assert(cfg.NoResetCredentials, check.Equals, expectedNoResetCredentials)
	c.Assert(cfg.NoOperational, check.Equals, expectedNoOperational)
}

func verifyDefaultLocalConfig(c *check.C, cfg *PortalConfig) {
	verifyLocalConfig(c, cfg, "", false, false)
}

func (s *S) TestReadLocalConfig(c *check.C) {
	f, err := createTempFile(testLocalConfig)
	c.Assert(err, check.IsNil)

	defer os.Remove(f.Name())
	configFile = f.Name()

	cfg, err := readLocalConfig()
	c.Assert(err, check.IsNil)

	verifyLocalConfig(c, cfg, "the_password", true, false)
}

func (s *S) TestReadLocalConfigBadEntry(c *check.C) {
	// No matter if there are additional not recognized params, only known should be marshalled
	f, err := createTempFile(testLocalConfigBadEntry)
	c.Assert(err, check.IsNil)

	defer os.Remove(f.Name())
	configFile = f.Name()

	cfg, err := readLocalConfig()
	c.Assert(err, check.IsNil)

	verifyLocalConfig(c, cfg, "the_password", true, false)
}

func (s *S) TestReadLocalEmptyConfig(c *check.C) {
	// No matter if there are additional not recognized params, only known should be marshalled
	f, err := createTempFile(testLocalEmptyConfig)
	c.Assert(err, check.IsNil)

	defer os.Remove(f.Name())
	configFile = f.Name()

	cfg, err := readLocalConfig()
	c.Assert(err, check.IsNil)

	verifyDefaultLocalConfig(c, cfg)
}

func (s *S) TestReadLocalNotExistingConfig(c *check.C) {
	configFile = "does/not/exists/config.json"

	cfg, err := readLocalConfig()
	c.Assert(err, check.IsNil)

	verifyDefaultLocalConfig(c, cfg)
}

func (s *S) TestWriteLocalConfigFileDoesNotExists(c *check.C) {
	mustConfigFlagFile = filepath.Join(os.TempDir(), "config_done"+randomName())
	defer os.Remove(mustConfigFlagFile)
	configFile = filepath.Join(os.TempDir(), "config"+randomName())
	defer os.Remove(configFile)

	c.Assert(MustSetConfig(), check.Equals, true)

	err := writeLocalConfig(testPortalConfig)
	c.Assert(err, check.IsNil)

	cfg, err := readLocalConfig()
	c.Assert(err, check.IsNil)

	c.Assert(*cfg, check.Equals, *testPortalConfig)
	c.Assert(MustSetConfig(), check.Equals, false)
}

func (s *S) TestWriteLocalConfigFiletExists(c *check.C) {
	mustConfigFlagFile = filepath.Join(os.TempDir(), "config_done"+randomName())
	defer os.Remove(mustConfigFlagFile)

	f, err := createTempFile(testLocalConfigBadEntry)
	c.Assert(err, check.IsNil)

	defer os.Remove(f.Name())
	configFile = f.Name()

	c.Assert(MustSetConfig(), check.Equals, true)

	err = writeLocalConfig(testPortalConfig)
	c.Assert(err, check.IsNil)

	cfg, err := readLocalConfig()
	c.Assert(err, check.IsNil)

	c.Assert(*cfg, check.Equals, *testPortalConfig)
	c.Assert(MustSetConfig(), check.Equals, false)
}

func (s *S) TestMustSetConfig(c *check.C) {
	mustConfigFlagFile = filepath.Join(os.TempDir(), "config_done"+randomName())
	defer os.Remove(mustConfigFlagFile)
	configFile = filepath.Join(os.TempDir(), "config"+randomName())
	defer os.Remove(configFile)

	c.Assert(MustSetConfig(), check.Equals, true)

	err := writeLocalConfig(testPortalConfig)
	c.Assert(err, check.IsNil)

	c.Assert(MustSetConfig(), check.Equals, false)

	err = writeLocalConfig(testPortalConfig)
	c.Assert(err, check.IsNil)

	c.Assert(MustSetConfig(), check.Equals, false)
}

// #####################
// Testing remote config
// #####################
type wifiapClientMock struct {
	m map[string]interface{}
}

func (c *wifiapClientMock) Show() (map[string]interface{}, error) {
	return c.m, nil
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

func (s *S) TestReadRemoteConfig(c *check.C) {
	wifiapClient = &wifiapClientMock{
		m: map[string]interface{}{
			"dhcp.lease-time":          "12h",
			"dhcp.range-start":         "10.0.60.2",
			"dhcp.range-stop":          "10.0.60.199",
			"disabled":                 true,
			"share.disabled":           false,
			"share.network-interface":  "wlp2s0",
			"wifi.address":             "10.0.60.1",
			"wifi.channel":             "6", // in real environment, channel is returned as string
			"wifi.hostapd-driver":      "nl80211",
			"wifi.interface":           "wlp2s0",
			"wifi.interface-mode":      "direct",
			"wifi.country-code":        "0x31",
			"wifi.netmask":             "255.255.255.0",
			"wifi.operation-mode":      "g",
			"wifi.security":            "wpa2",
			"wifi.security-passphrase": "17Soj8/Sxh14lcpD",
			"wifi.ssid":                "Ubuntu",
		},
	}

	cfg, err := readRemoteConfig()
	c.Assert(err, check.IsNil)

	expectedCfg := &WifiConfig{
		Ssid:          "Ubuntu",
		Passphrase:    "17Soj8/Sxh14lcpD",
		Interface:     "wlp2s0",
		CountryCode:   "0x31",
		Channel:       6,
		OperationMode: "g",
	}

	c.Assert(*cfg, check.Equals, *expectedCfg)
}

func (s *S) TestReadRemoteConfigNotAllParams(c *check.C) {
	wifiapClient = &wifiapClientMock{
		m: map[string]interface{}{
			"dhcp.lease-time":         "12h",
			"dhcp.range-start":        "10.0.60.2",
			"dhcp.range-stop":         "10.0.60.199",
			"share.disabled":          false,
			"share.network-interface": "wlp2s0",
			"wifi.address":            "10.0.60.1",
			"wifi.hostapd-driver":     "nl80211",
			"wifi.interface":          "wlp2s0",
			"wifi.interface-mode":     "direct",
			"wifi.country-code":       "0x31",
			"wifi.netmask":            "255.255.255.0",
			"wifi.security":           "wpa2",
			"wifi.ssid":               "Ubuntu",
		},
	}

	cfg, err := readRemoteConfig()
	c.Assert(err, check.IsNil)

	expectedCfg := &WifiConfig{
		Ssid:          "Ubuntu",
		Passphrase:    "",
		Interface:     "wlp2s0",
		CountryCode:   "0x31",
		Channel:       0,
		OperationMode: "",
	}

	c.Assert(*cfg, check.Equals, *expectedCfg)
}

func (s *S) TestReadEmptyRemoteConfig(c *check.C) {
	wifiapClient = &wifiapClientMock{}

	cfg, err := readRemoteConfig()
	c.Assert(err, check.IsNil)

	expectedCfg := &WifiConfig{
		Ssid:          "",
		Passphrase:    "",
		Interface:     "",
		CountryCode:   "",
		Channel:       0,
		OperationMode: "",
	}

	c.Assert(*cfg, check.Equals, *expectedCfg)
}

func (s *S) TestWriteRemoteConfig(c *check.C) {
	wifiapClient = &wifiapClientMock{}

	err := writeRemoteConfig(&WifiConfig{
		Ssid:          "Ubuntu",
		Passphrase:    "17Soj8/Sxh14lcpD",
		Interface:     "wlp2s0",
		CountryCode:   "0x31",
		Channel:       6,
		OperationMode: "g",
	})

	c.Assert(err, check.IsNil)
}

// ####################
// Testing whole config
// ####################
func (s *S) TestWriteConfig(c *check.C) {
	wifiapClient = &wifiapClientMock{}

	configFile = filepath.Join(os.TempDir(), "config"+randomName())
	defer os.Remove(configFile)

	mustConfigFlagFile = filepath.Join(os.TempDir(), "config_done"+randomName())
	defer os.Remove(mustConfigFlagFile)

	c.Assert(MustSetConfig(), check.Equals, true)

	err := WriteConfig(&Config{
		Wifi: &WifiConfig{
			Ssid:          "Ubuntu",
			Passphrase:    "17Soj8/Sxh14lcpD",
			Interface:     "wlp2s0",
			CountryCode:   "0x31",
			Channel:       6,
			OperationMode: "g",
		},
		Portal: &PortalConfig{
			Password:           "the_password",
			NoResetCredentials: true,
			NoOperational:      false,
		},
	})

	c.Assert(err, check.IsNil)

	c.Assert(MustSetConfig(), check.Equals, false)
}

func (s *S) TestReadConfig(c *check.C) {
	wifiapClient = &wifiapClientMock{
		m: map[string]interface{}{
			"dhcp.lease-time":          "12h",
			"dhcp.range-start":         "10.0.60.2",
			"dhcp.range-stop":          "10.0.60.199",
			"disabled":                 true,
			"share.disabled":           false,
			"share.network-interface":  "wlp2s0",
			"wifi.address":             "10.0.60.1",
			"wifi.channel":             "6", // in real environment, channel is returned as string
			"wifi.hostapd-driver":      "nl80211",
			"wifi.interface":           "wlp2s0",
			"wifi.interface-mode":      "direct",
			"wifi.country-code":        "0x31",
			"wifi.netmask":             "255.255.255.0",
			"wifi.operation-mode":      "g",
			"wifi.security":            "wpa2",
			"wifi.security-passphrase": "17Soj8/Sxh14lcpD",
			"wifi.ssid":                "Ubuntu",
		},
	}

	f, err := createTempFile(testLocalConfig)
	c.Assert(err, check.IsNil)

	defer os.Remove(f.Name())
	configFile = f.Name()

	expectedWifiConfig := &WifiConfig{
		Ssid:          "Ubuntu",
		Passphrase:    "17Soj8/Sxh14lcpD",
		Interface:     "wlp2s0",
		CountryCode:   "0x31",
		Channel:       6,
		OperationMode: "g",
	}

	expectedPortalConfig := &PortalConfig{
		Password:           "the_password",
		NoResetCredentials: true,
		NoOperational:      false,
	}

	cfg, err := ReadConfig()
	c.Assert(err, check.IsNil)
	c.Assert(*cfg.Wifi, check.Equals, *expectedWifiConfig)
	c.Assert(*cfg.Portal, check.Equals, *expectedPortalConfig)
}

func (s *S) TestConfigDump(c *check.C) {
	cfg := &Config{
		Wifi: &WifiConfig{
			Ssid:          "Ubuntu",
			Passphrase:    "17Soj8/Sxh14lcpD",
			Interface:     "wlp2s0",
			CountryCode:   "0x31",
			Channel:       6,
			OperationMode: "g",
		},
		Portal: &PortalConfig{
			Password:           "the_password",
			NoResetCredentials: true,
			NoOperational:      false,
		},
	}

	expected := `--Wifi--
Ssid:  Ubuntu
Passphrase: 17Soj8/Sxh14lcpD
Interface: wlp2s0
CountryCode: 0x31
Channel: 6
OperationMode: g
--Portal--
Password:  the_password
NoResetCredentials: true
NoOperational: false`

	c.Assert(cfg.String(), check.Equals, expected)
}

func (s *S) TestRollbackConfigIfFailsWriting(c *check.C) {
	//TODO IMPLEMENT
}

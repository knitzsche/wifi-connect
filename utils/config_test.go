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

func verifyLocalConfig(t *testing.T, cfg *PortalConfig, expectedPwd string, expectedNoResetCredentials bool, expectedNoOperational bool) {
	if cfg.Password != expectedPwd {
		t.Errorf("Local config portal.password is %v but expected %v", cfg.Password, expectedPwd)
	}

	if cfg.NoResetCredentials != expectedNoResetCredentials {
		t.Errorf("Local config portal.no-reset-creds is %v but expected %v", cfg.NoResetCredentials, expectedNoResetCredentials)
	}

	if cfg.NoOperational != expectedNoOperational {
		t.Errorf("Local config portal.no-operational is %v but expected %v", cfg.NoOperational, expectedNoOperational)
	}
}

func verifyDefaultLocalConfig(t *testing.T, cfg *PortalConfig) {
	verifyLocalConfig(t, cfg, "", false, false)
}

func TestReadLocalConfig(t *testing.T) {

	f, err := createTempFile(testLocalConfig)
	if err != nil {
		t.Errorf("Temp file error: %v", err)
	}

	defer os.Remove(f.Name())

	configFile = f.Name()

	cfg, err := readLocalConfig()
	if err != nil {
		t.Errorf("Error reading local config file: %v", err)
	}

	verifyLocalConfig(t, cfg, "the_password", true, false)
}

func TestReadLocalConfigBadEntry(t *testing.T) {

	// No matter if there are additional not recognized params, only known should be marshalled
	f, err := createTempFile(testLocalConfigBadEntry)
	if err != nil {
		t.Errorf("Temp file error: %v", err)
	}

	defer os.Remove(f.Name())

	configFile = f.Name()

	cfg, err := readLocalConfig()
	if err != nil {
		t.Errorf("Error reading local config file: %v", err)
	}

	verifyLocalConfig(t, cfg, "the_password", true, false)
}

func TestReadLocalEmptyConfig(t *testing.T) {

	// No matter if there are additional not recognized params, only known should be marshalled
	f, err := createTempFile(testLocalEmptyConfig)
	if err != nil {
		t.Errorf("Temp file error: %v", err)
	}

	defer os.Remove(f.Name())

	configFile = f.Name()

	cfg, err := readLocalConfig()
	if err != nil {
		t.Errorf("Error reading local config file: %v", err)
	}

	verifyDefaultLocalConfig(t, cfg)
}

func TestReadLocalNotExistingConfig(t *testing.T) {

	configFile = "does/not/exists/config.json"

	cfg, err := readLocalConfig()
	if err != nil {
		t.Errorf("Error reading local config file: %v", err)
	}

	verifyDefaultLocalConfig(t, cfg)
}

func TestWriteLocalConfigFileDoesNotExists(t *testing.T) {

	configFile = filepath.Join(os.TempDir(), "config"+randomName())
	defer os.Remove(configFile)

	err := writeLocalConfig(testPortalConfig)
	if err != nil {
		t.Errorf("Error while writting local config to file: %v", err)
	}

	cfg, err := readLocalConfig()
	if err != nil {
		t.Errorf("Error reading local config file: %v", err)
	}

	if *cfg != *testPortalConfig {
		t.Errorf("Got local config %v, but expected %v", cfg, testPortalConfig)
	}
}

func TestWriteLocalConfigFiletExists(t *testing.T) {

	f, err := createTempFile(testLocalConfigBadEntry)
	if err != nil {
		t.Errorf("Temp file error: %v", err)
	}

	defer os.Remove(f.Name())

	configFile = f.Name()

	err = writeLocalConfig(testPortalConfig)
	if err != nil {
		t.Errorf("Error while writting local config to file: %v", err)
	}

	cfg, err := readLocalConfig()
	if err != nil {
		t.Errorf("Error reading local config file: %v", err)
	}

	if *cfg != *testPortalConfig {
		t.Errorf("Got local config %v, but expected %v", cfg, testPortalConfig)
	}
}

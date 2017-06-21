package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

var testLocalConfig = `
{
	"portal.password": "the_password",
	"portal.no-reset-creds": true,
	"portal.no-operational": false	
}
`

var testLocalConfigBadEntry = `
{
	"portal.password": "the_password",
	"portal.no-reset-creds": true,
	"bad.parameter": "bad.value",
	"portal.no-operational": false
}
`

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

	if expected := "the_password"; cfg.Password != expected {
		t.Errorf("Local config portal.password is %v but expected %v", cfg.Password, expected)
	}

	if expected := true; cfg.NoResetCredentials != expected {
		t.Errorf("Local config portal.no-reset-creds is %v but expected %v", cfg.NoResetCredentials, expected)
	}

	if expected := false; cfg.NoOperational != expected {
		t.Errorf("Local config portal.no-operational is %v but expected %v", cfg.NoOperational, expected)
	}

}

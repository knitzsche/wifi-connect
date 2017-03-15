//
// Copyright (C) 2017 Canonical Ltd
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License version 3 as
// published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package wifiap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"path/filepath"
)

const (
	socketPath       = "/var/snap/wifi-ap/current/sockets/control"
	versionURI       = "/v1"
	configurationURI = "/configuration"
)

type serviceResponse struct {
	Result     map[string]interface{} `json:"result"`
	Status     string                 `json:"status"`
	StatusCode int                    `json:"status-code"`
	Type       string                 `json:"type"`
}

// RestClient defines client for rest api exposed by a unix socket
type RestClient struct {
	SocketPath string
	conn       net.Conn
}

// NewRestClient creates a RestClient object pointing to socket path set as parameter
func NewRestClient(socketPath string) *RestClient {
	return &RestClient{SocketPath: socketPath}
}

// DefaultRestClient created a RestClient object pointing to default socket path
func DefaultRestClient() *RestClient {
	return NewRestClient(socketPath) // FIXME this would be better like os.Getenv("SNAP_COMMON") + "/sockets/control", but that's not working
}

func (restClient *RestClient) unixDialer(_, _ string) (net.Conn, error) {
	return net.Dial("unix", restClient.SocketPath)
}

func (restClient *RestClient) sendHTTPRequest(uri string, method string, body io.Reader) (*serviceResponse, error) {
	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			Dial: restClient.unixDialer,
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	realResponse := &serviceResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&realResponse); err != nil {
		return nil, err
	}

	if realResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed: %s", realResponse.Result["message"])
	}

	return realResponse, nil
}

// Show renders current wifi-ap status
func (restClient *RestClient) Show() (map[string]interface{}, error) {
	uri := fmt.Sprintf("http://unix%s", filepath.Join(versionURI, configurationURI))
	response, err := restClient.sendHTTPRequest(uri, "GET", nil)
	if err != nil {
		return nil, err
	}

	return response.Result, nil
}

// Enable wifi-ap
func (restClient *RestClient) Enable() error {
	params := map[string]string{"disabled": "false"}
	b, err := json.Marshal(params)
	if err != nil {
		return err
	}

	uri := fmt.Sprintf("http://unix%s", filepath.Join(versionURI, configurationURI))
	response, err := restClient.sendHTTPRequest(uri, "POST", bytes.NewReader(b))
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK || response.Status != http.StatusText(http.StatusOK) {
		return fmt.Errorf("Failed to set configuration, service returned: %d (%s)\n", response.StatusCode, response.Status)
	}

	fmt.Println("Ok.")
	return nil
}

// Disable wifi-ap
func (restClient *RestClient) Disable() error {
	params := map[string]string{"disabled": "true"}
	b, err := json.Marshal(params)
	if err != nil {
		return err
	}

	uri := fmt.Sprintf("http://unix%s", filepath.Join(versionURI, configurationURI))
	response, err := restClient.sendHTTPRequest(uri, "POST", bytes.NewReader(b))
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK || response.Status != http.StatusText(http.StatusOK) {
		return fmt.Errorf("Failed to set configuration, service returned: %d (%s)\n", response.StatusCode, response.Status)
	}

	fmt.Println("Ok.")
	return nil
}

// SetSsid sets wifi SSID
func (restClient *RestClient) SetSsid(ssid string) error {
	params := map[string]string{"wifi.ssid": ssid}
	b, err := json.Marshal(params)
	if err != nil {
		return err
	}

	uri := fmt.Sprintf("http://unix%s", filepath.Join(versionURI, configurationURI))
	response, err := restClient.sendHTTPRequest(uri, "POST", bytes.NewReader(b))
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK || response.Status != http.StatusText(http.StatusOK) {
		return fmt.Errorf("Failed to set configuration, service returned: %d (%s)\n", response.StatusCode, response.Status)
	}

	fmt.Println("Ok.")
	return nil
}

// SetPassphrase sets wifi password
func (restClient *RestClient) SetPassphrase(passphrase string) error {
	params := map[string]string{
		"wifi.security":            "wpa2",
		"wifi.security-passphrase": passphrase,
	}
	b, err := json.Marshal(params)
	if err != nil {
		return err
	}

	uri := fmt.Sprintf("http://unix%s", filepath.Join(versionURI, configurationURI))
	response, err := restClient.sendHTTPRequest(uri, "POST", bytes.NewReader(b))
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK || response.Status != http.StatusText(http.StatusOK) {
		return fmt.Errorf("Failed to set configuration, service returned: %d (%s)\n", response.StatusCode, response.Status)
	}

	fmt.Println("Ok.")
	return nil
}

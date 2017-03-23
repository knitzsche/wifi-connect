package netman

import (
	"fmt"
	"os"
	"strings"

	"github.com/godbus/dbus"
)

func getDevices() []string {
	conn := getSystemBus()
	obj := conn.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager")
	var devices []string
	err2 := obj.Call("org.freedesktop.NetworkManager.GetAllDevices", 0).Store(&devices)
	if err2 != nil {
		panic(err2)
	}
	return devices
}

func getWifiDevices(devices []string) []string {
	conn := getSystemBus()
	var wifiDevices []string
	for _, d := range devices {
		objPath := dbus.ObjectPath(d)
		device := conn.Object("org.freedesktop.NetworkManager", objPath)
		deviceType, err2 := device.GetProperty("org.freedesktop.NetworkManager.Device.DeviceType")
		if err2 != nil {
			panic(err2)
		}
		var wifiType uint32
		wifiType = 2
		if deviceType.Value() == nil {
			break
		}
		if deviceType.Value() != wifiType {
			continue
		}
		wifiDevices = append(wifiDevices, d)
	}
	return wifiDevices
}

func getAccessPoints(devices []string, ap2device map[string]string) []string {
	conn := getSystemBus()
	var APs []string
	for _, d := range devices {
		objPath := dbus.ObjectPath(d)
		obj := conn.Object("org.freedesktop.NetworkManager", objPath)
		var aps []string
		err := obj.Call("org.freedesktop.NetworkManager.Device.Wireless.GetAllAccessPoints", 0).Store(&aps)
		if err != nil {
			panic(err)
		}
		if len(aps) == 0 {
			break
		}
		for _, i := range aps {
			APs = append(APs, i)
			ap2device[i] = d
		}
	}
	return APs
}

// SSID struct holds wireless SSID details
type SSID struct {
	Ssid   string
	ApPath string
}

func getSSIDs(APs []string, ssid2ap map[string]string) []SSID {
	conn := getSystemBus()
	var SSIDs []SSID
	for _, ap := range APs {
		objPath := dbus.ObjectPath(ap)
		obj := conn.Object("org.freedesktop.NetworkManager", objPath)
		ssid, err := obj.GetProperty("org.freedesktop.NetworkManager.AccessPoint.Ssid")
		if err != nil {
			panic(err)
		}
		type B []byte
		res := B(ssid.Value().([]byte))
		ssidStr := string(res)
		if len(ssidStr) < 1 {
			continue
		}
		found := false
		for _, s := range SSIDs {
			if s.Ssid == ssidStr {
				found = true
			}
		}
		if found == true {
			continue
		}

		Ssid := SSID{Ssid: ssidStr, ApPath: ap}
		SSIDs = append(SSIDs, Ssid)
		ssid2ap[strings.TrimSpace(ssidStr)] = ap
		//TODO: exclude ssid of device's own AP (the wifi-ap one)
	}
	return SSIDs
}

// ConnectAp connects to a specifid AP
func ConnectAp(ssid string, p string, ap2device map[string]string, ssid2ap map[string]string) {
	conn := getSystemBus()
	inner1 := make(map[string]dbus.Variant)
	inner1["security"] = dbus.MakeVariant("802-11-wireless-security")

	inner2 := make(map[string]dbus.Variant)
	inner2["key-mgmt"] = dbus.MakeVariant("wpa-psk")
	inner2["psk"] = dbus.MakeVariant(p)

	outer := make(map[string]map[string]dbus.Variant)
	outer["802-11-wireless"] = inner1
	outer["802-11-wireless-security"] = inner2
	fmt.Printf("%v\n", outer)

	fmt.Printf("dev path: %s\n", ap2device[ssid2ap[ssid]])
	fmt.Printf("ap path: %s\n", ssid2ap[ssid])

	obj := conn.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager")
	obj.Call("org.freedesktop.NetworkManager.AddAndActivateConnection", 0, outer, dbus.ObjectPath(ap2device[ssid2ap[ssid]]), dbus.ObjectPath(ssid2ap[ssid]))
}

func getSystemBus() *dbus.Conn {
	conn, err := dbus.SystemBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to system bus:", err)
		panic(1)
	}
	return conn
}

// Ssids returns an array of available ssids
func Ssids() ([]SSID, map[string]string, map[string]string) {
	ap2device := make(map[string]string)
	ssid2ap := make(map[string]string)

	devices := getDevices()
	wifiDevices := getWifiDevices(devices)
	APs := getAccessPoints(wifiDevices, ap2device)
	SSIDs := getSSIDs(APs, ssid2ap)
	return SSIDs, ap2device, ssid2ap
}

// ConnectedWifi returns true if connected to network by a netman interface
func ConnectedWifi() bool {
	conn := getSystemBus()
	objPath := dbus.ObjectPath("/org/freedesktop/NetworkManager")
	nm := conn.Object("org.freedesktop.NetworkManager", objPath)
	var nmConnectivityStatus uint32
	err := nm.Call("org.freedesktop.NetworkManager.CheckConnectivity", 0).Store(&nmConnectivityStatus)
	if err != nil {
		fmt.Printf("Error in ConnectedWifi(): %v\n", err)
		return false
	}
	if nmConnectivityStatus == 1 {
		return false //not connected by any netman interfaces
	}
	return true
}

// DisconnectWifi disconnect current network
func DisconnectWifi() {
	conn := getSystemBus()
	devices := getDevices()
	wifiDevices := getWifiDevices(devices)
	for _, d := range wifiDevices {
		objPath := dbus.ObjectPath(d)
		device := conn.Object("org.freedesktop.NetworkManager", objPath)
		device.Call("org.freedesktop.NetworkManager.Device.Disconnect", 0)
	}
	return
}

// SetIfaceManaged sets certain interface managed by netman
func SetIfaceManaged(iface string) {
	conn := getSystemBus()
	devices := getDevices()

	for _, d := range devices {
		objPath := dbus.ObjectPath(d)
		device := conn.Object("org.freedesktop.NetworkManager", objPath)
		deviceIface, err2 := device.GetProperty("org.freedesktop.NetworkManager.Device.Interface")
		if err2 != nil {
			fmt.Printf("Error in SetIfaceManaged(): %v\n", err2)
			return
		}
		if iface != deviceIface.Value().(string) {
			continue
		}
		managed, err := device.GetProperty("org.freedesktop.NetworkManager.Device.Managed")
		if err != nil {
			fmt.Printf("Error in SetIfaceManaged(): %v\n", err)
			return
		}
		if managed.Value().(bool) == true {
			return //no need to set as managed
		}

		device.Call("org.freedesktop.DBus.Properties.Set", 0, "org.freedesktop.NetworkManager.Device", "Managed", dbus.MakeVariant(true))
		managed, _ = device.GetProperty("org.freedesktop.NetworkManager.Device.Managed")

		return
	}
}

// WifisManaged list current netman managed wifis
func WifisManaged() map[string]string {
	conn := getSystemBus()
	devices := getDevices()
	wifiDevices := getWifiDevices(devices)

	ifaces := make(map[string]string)

	for _, d := range wifiDevices {
		objPath := dbus.ObjectPath(d)
		device := conn.Object("org.freedesktop.NetworkManager", objPath)
		managed, err := device.GetProperty("org.freedesktop.NetworkManager.Device.Managed")
		if err != nil {
			fmt.Printf("Error in wifiIfacesManaged(): %v\n", err)
			return ifaces
		}
		iface, err2 := device.GetProperty("org.freedesktop.NetworkManager.Device.Interface")
		if err2 != nil {
			fmt.Printf("Error in wifiIfacesManaged(): %v\n", err)
			return ifaces
		}
		if managed.Value().(bool) {
			ifaces[iface.Value().(string)] = d
		}
	}
	return ifaces
}
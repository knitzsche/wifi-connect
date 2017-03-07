package ssids

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

)

func args() *options {
	opts := &options{}
	flag.BoolVar(&opts.getSsids, "get-ssids", false, "Connect to an AP")
	flag.Parse()
	return opts
}

func doit() {
	opts := args()

	conn, err := dbus.SystemBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
		os.Exit(1)
	}

	ap2device := make(map[string]string)
	ssid2ap := make(map[string]string)

	devices := getDevices(conn)
	wifiDevices := getWifiDevices(conn, devices)
	APs := getAccessPoints(conn, wifiDevices, ap2device)
	SSIDs := getSSIDs(conn, APs, ssid2ap)
	if opts.getSsids {
		var out string
		for _, ssid := range SSIDs {
			out += strings.TrimSpace(ssid.ssid) + ","
		}
		fmt.Printf("%s\n", out[:len(out)-1])
		return
	}
	for _, ssid := range SSIDs {
		fmt.Printf("    %v\n", ssid.ssid)
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Connect to SSID: ")
	ssid, _ := reader.ReadString('\n')
	ssid = strings.TrimSpace(ssid)
	fmt.Print("PW: ")
	pw, _ := reader.ReadString('\n')
	pw = strings.TrimSpace(pw)
	connectAp(conn, ssid, pw, ap2device, ssid2ap)

	return
}

---
title: "Intallation"
table_of_contents: True
---

# Overview

The wifi-connect snap is currently publish in edge and beta channels.

## Install snaps

```bash
snap install wifi-ap
snap install network-manager
snap install --edge|beta wifi-connect
```

## Connect interfaces

```bash
snap connect wifi-connect:control wifi-ap:control
snap connect wifi-connect:network core:network
snap connect wifi-connect:network-bind core:network-bind
snap connect wifi-connect:network-manager network-manager:service
```

## Set NetWorkManager to control all networking

**Note**: This is a temporary manual step before network-manager snap provides a config option for this.

**Note**: Depending on your environment, after this you may need to use a new IP address to connect to the device.

 1. Backup the existing /etc/netplan/00-snapd-config.yaml file 

        sudo mv /etc/netplan/00-snapd-config.yaml ~/

 1. Create a new netplan config file named /etc/netplan/00-default-nm-renderer.yaml:

        sudo vi /etc/netplan/00-default-nm-renderer.yaml

    Add the following two lines:

        network:
            renderer: NetworkManager

## Reboot

Rebooting addresses a potential content sharing interface issue. 

Rebooting also consolidates all networking into NetworkManager.

## Optionally configure wifi-ap SSID/passphrase

If you skip these steps, the wifi-AP put up by the device has an SSID of "Ubuntu" and is unsecure (with no passphrase). 

 1. Set the wifi-ap AP SSID

        sudo  wifi-connect ssid MYSSID 

 1. Set the AP passphrase:

        sudo  wifi-connect passphrase MYPASSPHRASE

## Display the AP config

```bash
sudo  wifi-connect show-ap
```

**Note** the DHCP range:

    dhcp.range-start: 10.0.60.2
    dhcp.range-stop: 10.0.60.199

## Set the portal password

The portal password must be entered to access wifi-connect web pages.

```bash
sudo  wifi-connect set-portal-password PASSWORD
```

## Join the device AP

When the device AP is up and available to you, join it.

## Open the the Management portal web page

This portal displays external wifi APs and let's you join them.

After you connect to the device AP, you can open its http portal at the .1 IP address just before the start of the DHCP range (see previous steps) using port 8080: 

    10.0.60.1:8080

You then need to enter the portal password to continue.

### Avahi and hostname

You can also connect to the device's web page using the device host name: 

    http://HOSTNAME.local:8080 

Where HOSTNAME is the hostname of the device when it booted. (Changing hostname with the hostname command at run time is not sufficient.) 

**Note**: The system trying to open the web page must support Avahi. Android systems may not, for example.

## Be patient, it takes minutes

Wifi-connect pauses for about a minute at daemon start to allow any external AP connections to complete.

## Disconnect from wifi

When connected to an external AP, the Operational portal is available on the device IP address (assigned by the external AP). Open it using IP:8080, enter the portal password, and you may then disconnect with the "Disconnect from Wifi" button.

You can also ssh to the device and:

 * Use `nmcli c` to display connections.
 * Use `nmcli c delete CONNECTION_NAME` to disconnect and delete. This puts the device into management mode, bringing up the AP and portal.

Disconnecting sets the device back in Management mode. Its AP is started and you can open the portal (as discussed above) to see external APs and connect to one.


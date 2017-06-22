---
title: "Overview"
table_of_contents: True
---

# Overview

The wifi-connect snap allows you to connect your device to an external Wi-Fi access point. It does this by putting up its own Wi-Fi AP and web page. You join the AP and the web page lists external APs, which you can then select and join. 

Wifi-connect is appropriate for simple use cases where there is no other control of networking. Wifi-connect has a daemon that takes over networking and controls device state automatically:

 * When there is no external AP connection, wifi-connect starts its own AP and the Management web page (which allows you to select and join external WiFI APs)
 * When there is an external AP connection, wifi-connect ensures its own AP is down and puts up the Operational web page (which allows you to disconnect from the external AP) 

Wifi-connect uses two other snaps:

 * wifi-ap: provides the AP function
 * network-manager: handles networking (as a part of installation, the device netplan is modified to make network-manager the renderer for all networking).

Wifi-connect can be:

 * Installed at run time
 * Integrated into an image, with options. See "Integrating into an Image" section


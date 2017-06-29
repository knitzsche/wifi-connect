---
title: "Integrating into an image""
table_of_contents: True
---

# Overview

When you pre-install wifi-connect snap into an image, you can use the gadget snap's [gadget.yaml](https://forum.snapcraft.io/t/the-gadget-snap/696) file to pre-configure some options.

Here we explain the snap key and value needed to preconfigure.  Refer to the above like for details on how to set these in the gadget snap's gadget.yaml file. 

You may also set these at run time from terminal with:

```bash
$ snap set wifi-connect KEY=VALUE
```

To apply such run-time changes, see the Frequently asked questions page.

**Warning**: These changes may create security risks. Only use take these steps if you are completely aware of take responsitbility for the potential risk.


**Note** When the deamon starts, it logs any preconfigurations found and applied, for example:

```bash
Jun 20 22:07:16 thehost snap[18004]: == wifi-connect/SetDefaults portal password being set
Jun 20 22:07:16 thehost snap[18004]: == wifi-connect/SetDefaults: reset creds requirement is now disabled
```

## AP passphrase

Normally the wifi-conect AP's passphrase is randomly created by the wifi-ap snap. A normal part of the installation process is resetting this from the terminal. However, some integrators may want to preset the passphrase. 

Preset the passphrase with the following:

 * snapd key: wifi.security-passphrase
 * value: the passhrase (8-13 characters, must start with a letter)

## Portal password 

To access any wifi-connect web page you need to enter the portal password. A normal part of the installation process is setting this from the terminal. However, some integrators may want to preset the portal password.  

Preset the portal password  with the following:

 * snapd key: portal.password
 * value: the password (8-13 characters, must start with a letter)

## Disable credential resetting

Normally, the first user to access the Management portal is required to reset the wifi-connect AP passphrase and the portal password (used to access wifi-connect web pages). This is an important security feature, especially for integrators that preset these in their image because every image has the same passphrase and password. 

However, some integrators may want to disable this requirement. 

Disable the requirement to reset the AP passphrase and portal password on first use of the Management portal as follows:
 
 * snapd key: portal.no-reset-creds
 * value: true

## Disable display of the Operational Portal

The Operational portal is available when the device is connected to an external AP. It provides a button the user can click to disconnect. Display of this page can be disabled. This is a normal part of wifi-connect. When the page is disabled one must use the terminal (or some other means) to direct the device to disconnect form the external AP.

Disable the Operational portal as follows:

 * snapd key: portal.no-opertaional
 * value: true





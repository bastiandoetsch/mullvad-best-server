# mullvad-best-server
![Build](https://github.com/bastiandoetsch/mullvad-best-server/actions/workflows/go.yml/badge.svg)

Determines the mullvad.net server with the lowest latency. 

## Installation

Download binary from releases for your platform and unpack.

## Usage
### Default usage
Execute `mullvad-best-server`. It outputs the code, e.g. `de05`. You can then connect to it with e.g. wireguard using the normal shell scripts.

### Command line parameters

```angular2html
Usage of dist/mullvad-best-server_darwin_amd64/mullvad-best-server:
-c string
  Server country code, e.g. ch for Switzerland (default "ch")
-l string
  Log level. Allowed values: trace, debug, info, warn, error, fatal, panic (default "info")
-o string
  Output format. 'json' outputs server json
-t string
  Server type, e.g. wireguard (default "wireguard")
```

If you want the full server information, execute `mullvad-best-server -o json`. It returns the full json output  of the server information.
The `-c` flag allows to give a country code. Else `ch` will be used.


## Background
The program uses `https://api.mullvad.net/www/relays/<SERVER_TYPE>/` to get the current server list, pings the ones with the right country
and outputs the server with the lowest ping.

## Integration into a script
I use it on my router like this (yes, I know I could have done the whole thing with jq and shell scripting, but wanted to use go for maintainability).
```
#!/bin/sh
set -e
LATEST_RELEASE=$(curl -sSL https://api.github.com/repos/bastiandoetsch/mullvad-best-server/releases/latest | jq -r '.assets[]| .browser_download_url' | grep Linux_arm64)
curl -sSL $LATEST_RELEASE > /root/mullvad-best-server
chmod +x /root/mullvad-best-server
/usr/bin/wg-quick down $(wg show|grep interface | cut -d: -f2)  || echo "nothing to shut down"
/usr/bin/wg-quick up "mullvad-$(/root/mullvad-best-server -c de)"
```

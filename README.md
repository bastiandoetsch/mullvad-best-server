# mullvad-best-server
![Build](https://github.com/bastiandoetsch/mullvad-best-server/actions/workflows/go.yml/badge.svg)

Determines the mullvad.net server with the lowest latency on macOS and Linux. On Windows, it can only check if the server is up.
The reason lies with the golang network libraries, according to the `go-ping` library, that is used under the hood for pinging: 

```
Please note that accessing packet TTL values is not supported due to limitations in the Go x/net/ipv4 and x/net/ipv6 packages.
```

## Installation

Download binary from releases for your platform and unpack.

## Usage
### Default usage
Execute `mullvad-best-server`. It outputs the code, e.g. `de05`. You can then connect to it with e.g. wireguard using the normal shell scripts.

### Command line parameters

```angular2html
Usage of ./mullvad-best-server:
  -c string
    	Server country code, e.g. ch for Switzerland (default "ch")
  -l string
    	Log level. Allowed values: trace, debug, info, warn, error, fatal, panic (default "info")
  -o string
    	Output format. 'json' outputs server json
  -p string
    	filter by provider, e.g. 31173 for mullvad-owned
  -t string
    	Server type, e.g. wireguard (default "wireguard")
  -timeout duration
    	Timeout for network calls as duration, e.g. 60s (default 5s)
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

## Troubleshooting
You may have to update your `/usr/bin/wg-quick` script for the new mullvad servers, as they have names longer than 15 chars. Replace `[[ $CONFIG_FILE =~ ^[a-zA-Z0-9_=+.-]{1,15} ]]$` with `[[ $CONFIG_FILE =~ ^[a-zA-Z0-9_=+.-]{1,25}$ ]]` and `[[ $CONFIG_FILE =~ (^|/)([a-zA-Z0-9_=+.-]{1,15})\.conf$ ]]` with `[[ $CONFIG_FILE =~ (^|/)([a-zA-Z0-9_=+.-]{1,25})\.conf$ ]]`

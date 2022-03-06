# mullvad-best-server
Determines the mullvat.net wireguard server with the lowest latency. 

## Installation

Download binary from releases for your platform and unpack.

## Usage
### Default usage
Execute `mullvad-best-server`. It outputs the code, e.g. `de05`. You can then connect to it with e.g. wireguard using the normal shell scripts.

### Extended Output
If you want the full server information, execute `mullvad-best-server -o json`. It returns the full json output  of the server information. 

## Background
The program uses `https://api.mullvad.net/www/relays/wireguard/` to get the current server list, pings the ones with the right country
and outputs the server with the lowest ping.

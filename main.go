package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/go-ping/ping"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	var outputFlag = flag.String("o", "", "Output format. 'json' outputs server json")
	var countryFlag = flag.String("c", "ch", "Server country code, e.g. ch for Switzerland")
	var typeFlag = flag.String("t", "wireguard", "Server type, e.g. wireguard")
	var logLevel = flag.String("l", "info", "Log level. Allowed values: trace, debug, info, warn, error, fatal, panic")
	var timeout = flag.Duration("timeout", time.Second*5, "Timeout for network calls as duration, e.g. 60s")
	var provider = flag.String("p", "", "filter by provider, e.g. 31173 for mullvad-owned")
	flag.Parse()

	level, err := zerolog.ParseLevel(*logLevel)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to set log level")
	}
	zerolog.SetGlobalLevel(level)
	servers := getServers(*typeFlag, *timeout)
	bestIndex := selectBestServerIndex(servers, *countryFlag, *provider)
	if bestIndex == -1 {
		log.Fatal().Str("country", *countryFlag).Msg("No servers for country found.")
	}
	best := servers[bestIndex]
	log.Debug().Interface("server", best).Msg("Best latency server found.")
	hostname := strings.TrimSuffix(best.Hostname, "-wireguard")
	if *outputFlag != "json" {
		fmt.Println(hostname)
	} else {
		serverJson, err := json.Marshal(best)
		if err != nil {
			log.Fatal().Err(err).Msg("Couldn't marshal server information to Json")
		}
		fmt.Println(string(serverJson))
	}
}

func selectBestServerIndex(servers []server, country string, provider string) int {
	bestIndex := -1
	var bestPing time.Duration
	for i, server := range servers {
		if server.Active && server.CountryCode == country && (provider == "" || server.Provider == provider) {
			duration, err := serverLatency(server)
			if err == nil {
				if bestIndex == -1 || bestPing > duration {
					bestIndex = i
					bestPing = duration
				}
			} else {
				log.Err(err).Msg("Error determining the server latency via ping.")
			}
		}
	}
	return bestIndex
}

func getServers(serverType string, timeout time.Duration) []server {
	client := http.DefaultClient
	client.Timeout = timeout
	resp, err := client.Get("https://api.mullvad.net/www/relays/" + serverType + "/")
	if err != nil {
		log.Fatal().Err(err).Msg("Couldn't retrieve servers")
	}
	responseBody, err := io.ReadAll(resp.Body)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Err(err)
		}
	}(resp.Body)
	if err != nil {
		log.Fatal().Err(err)
	}
	var servers []server
	err = json.Unmarshal(responseBody, &servers)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't unmarshall server json")
	}
	return servers
}

//goland:noinspection GoBoolExpressions
func serverLatency(s server) (time.Duration, error) {
	pinger, err := ping.NewPinger(s.Ipv4AddrIn)
	if runtime.GOOS == "windows" {
		pinger.SetPrivileged(true)
	}
	pinger.Count = 1
	pinger.Timeout = time.Millisecond * 100
	if err != nil {
		return 0, err
	}
	var duration time.Duration
	pinger.OnRecv = func(pkt *ping.Packet) {
		log.Debug().Str("Server", s.Hostname).IPAddr("IP", pkt.IPAddr.IP).Dur("RTT", pkt.Rtt).Msg("Added server latency.")
		duration = pkt.Rtt
	}
	err = pinger.Run()
	return duration, err
}

type server struct {
	Hostname         string `json:"hostname"`
	CountryCode      string `json:"country_code"`
	CountryName      string `json:"country_name"`
	CityCode         string `json:"city_code"`
	CityName         string `json:"city_name"`
	Active           bool   `json:"active"`
	Owned            bool   `json:"owned"`
	Provider         string `json:"provider"`
	Ipv4AddrIn       string `json:"ipv4_addr_in"`
	Ipv6AddrIn       string `json:"ipv6_addr_in"`
	NetworkPortSpeed int    `json:"network_port_speed"`
	Pubkey           string `json:"pubkey"`
	MultihopPort     int    `json:"multihop_port"`
	SocksName        string `json:"socks_name"`
}

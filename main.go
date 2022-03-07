package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/go-ping/ping"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	var outputFlag = flag.String("o", "", "Output format. 'json' outputs server json")
	var countryFlag = flag.String("c", "ch", "Server country code, e.g. ch for Switzerland")
	var typeFlag = flag.String("t", "wireguard", "Server type, e.g. wireguard")
	flag.Parse()

	servers := getServers(*typeFlag)
	bestIndex := selectBestServerIndex(servers, *countryFlag)
	best := servers[bestIndex]
	log.Debug().Interface("server", best).Msg("Best latency server found.")
	hostname := strings.TrimSuffix(best.Hostname, "-wireguard")
	if *outputFlag != "json" {
		fmt.Println(hostname)
	} else {
		serverJson, err := json.Marshal(best)
		if err != nil {
			log.Fatal().Err(err)
		}
		fmt.Println(string(serverJson))
	}
}

func selectBestServerIndex(servers []server, country string) int {
	bestIndex := -1
	var bestPing time.Duration
	for i, server := range servers {
		if server.Active && server.CountryCode == country {
			duration, err := serverLatency(server)
			if err == nil {
				if bestIndex == -1 || bestPing > duration {
					bestIndex = i
					bestPing = duration
				}
			}
		}
	}
	return bestIndex
}

func getServers(serverType string) []server {
	resp, err := http.Get("https://api.mullvad.net/www/relays/" + serverType + "/")
	if err != nil {
		log.Fatal().Err(err)
	}
	responseBody, err := ioutil.ReadAll(resp.Body)
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
		log.Fatal().Err(err)
	}
	return servers
}

func serverLatency(s server) (time.Duration, error) {
	pinger, err := ping.NewPinger(s.Ipv4AddrIn)
	pinger.Count = 1
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

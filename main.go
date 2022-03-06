package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-ping/ping"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

var pings = make(map[string]time.Duration)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	servers := getServers()
	bestIndex := selectBestServerIndex(servers)
	log.Debug().Interface("server", servers[bestIndex]).Msg("Best latency server found.")
	fmt.Println(servers[bestIndex].Hostname)
}

func selectBestServerIndex(servers []server) int {
	best := servers[0].Hostname
	bestIndex := 0
	allowedCountries := map[string]string{}
	allowedCountries["de"] = "1"
	allowedCountries["ch"] = "1"
	allowedCountries["at"] = "1"
	for i, server := range servers {
		if server.Active && allowedCountries[server.CountryCode] != "" {
			duration, err := serverLatency(server)
			if err == nil {
				pings[server.Hostname] = duration
				if best == "" || pings[best] > pings[server.Hostname] {
					best = server.Hostname
					bestIndex = i
				}
			}
		}
	}
	return bestIndex
}

func getServers() []server {
	resp, err := http.Get("https://api.mullvad.net/www/relays/wireguard/")
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

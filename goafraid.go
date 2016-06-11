// goafraid - freedns.afraid.org updater in Go.

// Copyright 2016 Diego Elio Pettenò <flameeyes@flameeyes.eu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"io/ioutil"
)

var (
	configFile = flag.String("config_file", "", "Path to the JSON configuration file with the settings.")
	endpoint   = url.URL{
		Scheme: "https",
		Host:   "sync.afraid.org",
		Path:   "/u/",
	}
	httpClient = &http.Client{}
)

const (
	familyV4 = "ip4"
	familyV6 = "ip6"
)

type Host struct {
	Name          string
	Interface     string
	AddressFamily string
}

type Config struct {
	User     string
	Password string
	Hosts    []Host
}

func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	d := json.NewDecoder(f)
	cfg := new(Config)
	if err := d.Decode(cfg); err != nil {
		return nil, err
	}

	if cfg.User == "" || cfg.Password == "" {
		return nil, fmt.Errorf("Missing User or Password value.")
	}

	for _, h := range cfg.Hosts {
		h.AddressFamily = strings.ToLower(h.AddressFamily)
		if h.AddressFamily == "" {
			h.AddressFamily = familyV4
		}
	}

	return cfg, nil
}

func BuildUpdateURL(userEndpoint url.URL, host Host) (*url.URL, error) {
	qv := url.Values{}
	qv.Set("content-type", "json")
	qv.Set("h", host.Name)

	iface, err := net.InterfaceByName(host.Interface)
	if err != nil {
		return nil, err
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}

	found := false
	for _, addr := range addrs {
		switch addr.(type) {
		case *net.IPNet:
			ipAddr := addr.(*net.IPNet).IP
			switch host.AddressFamily {
			case familyV4:
				if ipAddr.To4() != nil {
					found = true
					qv.Set("ip", ipAddr.String())
				}
			case familyV6:
				if ipAddr.To4() == nil && ipAddr.IsGlobalUnicast() {
					found = true
					qv.Set("ip", ipAddr.String())
				}
			}
		default:
			return nil, fmt.Errorf("Unexpected address type %v for interface %v", addr.Network(), host.Interface)
		}
	}

	if !found {
		return nil, fmt.Errorf("Unable to find address of family %v on interface %v", host.AddressFamily, host.Interface)
	}

	hostEndpoint := userEndpoint
	hostEndpoint.RawQuery = qv.Encode()

	return &hostEndpoint, nil
}

func UpdateHost(userEndpoint url.URL, host Host) error {
	hostEndpoint, err := BuildUpdateURL(userEndpoint, host)
	if err != nil {
		return err
	}

	resp, err := http.Get(hostEndpoint.String())
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Non-OK status received: %v", resp.Status)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	log.Printf("%s", body)

	return nil
}

func main() {
	flag.Parse()

	log.Println("goafraid - afraid.org updater by Diego Elio Pettenò <flameeyes@flameeyes.eu>")
	if *configFile == "" {
		log.Fatalf("Missing value for -config_file.")
	}

	cfg, err := LoadConfig(*configFile)
	if err != nil {
		log.Fatalln(err)
	}

	userEndpoint := endpoint
	userEndpoint.User = url.UserPassword(cfg.User, cfg.Password)

	for _, host := range cfg.Hosts {
		err := UpdateHost(userEndpoint, host)
		if err != nil {
			log.Fatal(err)
		}
	}
}

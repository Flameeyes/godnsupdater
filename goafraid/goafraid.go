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
	"github.com/flameeyes/godnsupdater"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

var (
	configFile = flag.String("config_file", "", "Path to the JSON configuration file with the settings.")
	httpClient = &http.Client{}

	// Make sure that we use the v6 URL when updating a v6 address. This might not be the perfect assumption, but if we hit the non-v6 address we might actually not have a network to get to.
	familyEndpoints = map[godnsupdater.AddressFamily]string{
		godnsupdater.IPv4: "https://sync.afraid.org/u/",
		godnsupdater.IPv6: "https://v6.sync.afraid.org/u/",
	}
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
		if _, err := godnsupdater.FamilyFromString(h.AddressFamily); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

func BuildUpdateURL(user *url.Userinfo, host Host) (string, error) {
	family, err := godnsupdater.FamilyFromString(host.AddressFamily)
	if err != nil {
		return "", err
	}

	address, err := godnsupdater.GetInterfaceIP(host.Interface, family)
	if err != nil {
		return "", err
	}

	qv := url.Values{}
	qv.Set("content-type", "json")
	qv.Set("h", host.Name)
	qv.Set("ip", address)

	endpoint, err := url.Parse(familyEndpoints[family])
	if err != nil {
		return "", err
	}

	endpoint.RawQuery = qv.Encode()
	endpoint.User = user

	return endpoint.String(), nil
}

func UpdateHost(user *url.Userinfo, host Host) error {
	endpoint, err := BuildUpdateURL(user, host)
	if err != nil {
		return err
	}

	resp, err := http.Get(endpoint)
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

	user := url.UserPassword(cfg.User, cfg.Password)

	for _, host := range cfg.Hosts {
		err := UpdateHost(user, host)
		if err != nil {
			log.Fatal(err)
		}
	}
}

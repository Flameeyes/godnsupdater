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
	"flag"
	"fmt"
	"github.com/flameeyes/godnsupdater"
	"github.com/prasmussen/gandi-api/client"
	"github.com/prasmussen/gandi-api/domain/zone"
	"github.com/prasmussen/gandi-api/domain/zone/record"
	"github.com/prasmussen/gandi-api/domain/zone/version"
	"io/ioutil"
	"log"
	"strings"
)

var (
	apiFilePath           = flag.String("api_file", "", "Path to the file containing the Gandi API key to use")
	useTestingEnvironment = flag.Bool("use_ote", false, "Whether to use the Gandi Testing Environment")

	ifaceName  = flag.String("iface", "", "Name of the local interface to get the address from")
	addrFamily = flag.String("family", "ip4", "Address family for the IP to set")
	zoneId     = flag.Int64("zone", 0, "ID of the Gandi zone to use.")
	recordName = flag.String("record", "", "Name of the record to use (local part of the hostname)")
)

func getApiKey(path string) (string, error) {
	api, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("Error reading API key from \"%v\": %v", path, err)
	}

	return strings.Trim(string(api), "\n "), nil
}

func cloneLatestZone(c *client.Client, zId int64) (int64, error) {
	z := zone.New(c)
	zInfo, err := z.Info(zId)
	if err != nil {
		return 0, nil
	}

	zv := version.New(c)
	newVersion, err := zv.New(zId, zInfo.Version)
	if err != nil {
		return 0, nil
	}

	return newVersion, nil
}

func main() {
	flag.Parse()
	log.Println("gogogandidns - Gandi Dynamic DNS updater by Diego Elio Pettenò <flameeyes@flameeyes.eu>")

	if *apiFilePath == "" {
		log.Fatalf("Missing value for -api_file")
	}

	if *ifaceName == "" {
		log.Fatalf("Missing value for -iface")
	}

	family, err := godnsupdater.FamilyFromString(*addrFamily)
	if err != nil {
		log.Fatalf("Invalid value for -family: %v", err)
	}

	if *zoneId == 0 {
		log.Fatalf("Missing value for -zone")
	}

	if *recordName == "" {
		log.Fatalf("Missing value for -record")
	}

	api, err := getApiKey(*apiFilePath)
	if err != nil {
		log.Fatal(err)
	}

	address, err := godnsupdater.GetInterfaceIP(*ifaceName, family)
	if err != nil {
		log.Fatal(err)
	}

	env := client.Production
	if *useTestingEnvironment {
		env = client.Testing
	}
	c := client.New(api, env)

	newVersion, err := cloneLatestZone(c, *zoneId)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("New version %v created for zone %v", newVersion, *zoneId)

	recordType := godnsupdater.DnsTypeByFamily[family]

	// Get the list of current entries, filter on the name and type of record, as delete all the records of the same type already present.
	r := record.New(c)
	rInfos, err := r.List(*zoneId, newVersion)
	if err != nil {
		log.Fatal(err)
	}
	for _, rInfo := range(rInfos) {
		if rInfo.Name == *recordName && rInfo.Type == recordType {
			log.Printf("Removing record \"%v\" with ID %v", rInfo.Name, rInfo.Id)
			ok, err := r.Delete(*zoneId, newVersion, rInfo.Id)
			if err != nil {
				log.Fatal(err)
			}
			if !ok {
				log.Fatalf("Deleting record failed, but no error returned.")
			}
		}
	}
	
	addArgs := record.RecordAdd{
		Zone:    *zoneId,
		Version: newVersion,
		Name:    *recordName,
		Type:    recordType,
		Value:   address,
		Ttl:     300,
	}
	rInfo, err := r.Add(addArgs)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("New record \"%v\" created with ID %v", rInfo.Name, rInfo.Id)

	zv := version.New(c)
	ok, err := zv.Set(*zoneId, newVersion)
	if err != nil {
		log.Fatal(err)
	}
	if !ok {
		log.Fatalf("Setting new version live failed, but no error returned.")
	}
	log.Printf("Version %v set live for zone %v", newVersion, *zoneId)
}

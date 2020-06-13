// Copyright 2016 Diego Elio Pettenò <flameeyes@flameeyes.com>
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

package godnsupdater

import (
	"fmt"
	"strings"
)

const (
	Unknown AddressFamily = iota
	IPv4
	IPv6
)

var (
	DnsTypeByFamily = map[AddressFamily]string{
		IPv4: "A",
		IPv6: "AAAA",
	}
)

type AddressFamily int

func FamilyFromString(family string) (AddressFamily, error) {
	switch strings.ToLower(family) {
	case "ip4", "ipv4", "inet4", "4":
		return IPv4, nil
	case "ip6", "ipv6", "inet6", "6":
		return IPv6, nil
	}

	return Unknown, fmt.Errorf("Unable to parse \"%v\" as an address family", family)
}

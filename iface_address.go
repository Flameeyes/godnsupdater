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

package godnsupdater

import (
	"fmt"
	"net"
)

func GetInterfaceIP(ifaceName string, family AddressFamily) (string, error) {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return "", err
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		switch addr.(type) {
		case *net.IPNet:
			ipAddr := addr.(*net.IPNet).IP
			switch family {
			case IPv4:
				if ipAddr.To4() != nil {
					return ipAddr.String(), nil
				}
			case IPv6:
				if ipAddr.To4() == nil && ipAddr.IsGlobalUnicast() {
					return ipAddr.String(), nil
				}
			}
		default:
			return "", fmt.Errorf("Unexpected address type %v for interface %v", addr.Network(), ifaceName)
		}
	}

	return "", fmt.Errorf("Unable to find address of family %v on interface %v", family, ifaceName)
}

# godnsupdater

A set of dynamic DNS updaters in Go.

Released under the terms of [Apache License, Version 2.0][licence]

## Background

I started this as a way to updater my afraid.org FreeDNS entry, since they
support AAAA (IPv6) records. Since I then wanted to support other providers
(such as Gandi) I started factoring out common functions into a library and a
set of tools.

## Author

Diego Elio Pettenò <flameeyes@flameeyes.com>

[licence]: http://www.apache.org/licenses/LICENSE-2.0

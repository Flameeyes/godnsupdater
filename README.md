# goafraid

A freedns.afraid.org updater in Go.

Released under the terms of [Apache License, Version 2.0][licence]

## Background

afraid.org provides one of the few free Dynamic DNS services supporting AAAA
(IPv6) records. This makes it very valuable for making services accessible when
running on dynamic IPv6 addresses, such as [DS-Lite][dslite] users.

A simple `curl`-able API is available, but a client can have some extra
features, unfortunately most suggested clients go the minimal route, or don't
implement IPv6 support.

## Future plans

While there is no described API to update TXT records, I've asked the service
owner about it. If that is available, supporting [dns-01 challenge][dns-01] for
the ACME protocol would make it the perfect client for generating *Let's
Encrypt* certificates, while their service still does not support IPv6-only
hosts.

## Author

Diego Elio Pettenò <flameeyes@flameeyes.eu>

[licence]: http://www.apache.org/licenses/LICENSE-2.0
[dslite]: https://en.wikipedia.org/wiki/DS-Lite
[dns-01]: https://tools.ietf.org/html/draft-ietf-acme-acme-01#section-7.5

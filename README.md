# geoiplookup

GeoIP lookup utility. Required MaxMind MMDB databases.

## Install

Download binaries from [release](https://github.com/n0madic/geoiplookup/releases) page.

Or install from source:

```
go install github.com/n0madic/geoiplookup@latest
```

## Help

```
Usage: geoiplookup [--asn PATH] [--geo PATH] [--lang LANG] [--path PATH] [--whois] TARGET

Positional arguments:
  TARGET                 IP or domain for lookup

Options:
  --asn PATH, -a PATH    MaxMind ASN database path (optional) [default: GeoLite2-ASN.mmdb]
  --geo PATH, -g PATH    MaxMind GeoIP2 database path
  --lang LANG, -l LANG   MaxMind GeoIP2 database language [default: en]
  --path PATH, -p PATH   Path prefix to MaxMind databases
  --whois, -w            Lookup Whois information
  --help, -h             display this help and exit
```

## Usage

```
$ geoiplookup 8.8.8.8
┌────────────────────────────────────────────────────────────────────────────┐
│                                 GEOIP INFO                                 │
├───────────────────┬────────────────────────────────────────────────────────┤
│ IP:               │ 8.8.8.8 <dns.google>                                   │
│ Continent:        │ North America                                          │
│ Country:          │ United States                                          │
│ ISO code:         │ US                                                     │
│ Timezone:         │ America/Chicago                                        │
│ Coordinates:      │ 37.751000,-97.822000                                   │
│ Google Map URL:   │ https://www.google.com/maps/place/37.751000,-97.822000 │
│ ASN Number:       │ 15169                                                  │
│ ASN Organization: │ GOOGLE                                                 │
└───────────────────┴────────────────────────────────────────────────────────┘
```

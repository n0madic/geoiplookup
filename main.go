package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/likexian/whois"
	"github.com/oschwald/geoip2-golang"
	"golang.org/x/term"
	_ "golang.org/x/term"
)

var (
	args struct {
		ASN    string   `arg:"-a" placeholder:"PATH" default:"GeoLite2-ASN.mmdb" help:"MaxMind ASN database path (optional)"`
		GeoDB  []string `arg:"-g,--geo" placeholder:"PATH" help:"MaxMind GeoIP2 database path"`
		Target string   `arg:"positional,required" help:"IP or domain for lookup"`
		Lang   string   `arg:"-l" placeholder:"LANG" default:"en" help:"MaxMind GeoIP2 database language" default:"en"`
		Path   string   `arg:"-p" help:"Path prefix to MaxMind databases"`
		Whois  bool     `arg:"-w" help:"Lookup Whois information"`
	}

	geoIP2Filenames = []string{
		"GeoIP2-City.mmdb",
		"GeoLite2-City.mmdb",
		"GeoIP2-Country.mmdb",
		"GeoLite2-Country.mmdb",
	}
)

type geoDB struct {
	GeoIP2 *geoip2.Reader
	ASN    *geoip2.Reader
}

func main() {
	var err error
	var db geoDB

	arg.MustParse(&args)
	args.Lang = strings.ToLower(args.Lang)

	// If the path is not specified, then search in standard directories
	if args.Path == "" && len(args.GeoDB) == 0 {
		searchPath := []string{
			".",
			"/usr/share/GeoIP/",
			"/usr/local/share/GeoIP/",
			"/var/lib/GeoIP/",
			"/opt/homebrew/var/GeoIP",
		}
		for _, path := range searchPath {
			matches, err := filepath.Glob(filepath.Join(path, "*.mmdb"))
			if len(matches) > 0 && err == nil {
				args.Path = path
				break
			}
		}
	}

	// Checking the slash at the end of the path
	if args.Path != "" && !strings.HasSuffix(args.Path, string(os.PathSeparator)) {
		args.Path += string(os.PathSeparator)
	}

	// Default list of database file names
	if len(args.GeoDB) == 0 {
		args.GeoDB = geoIP2Filenames
	}

	// Trying to open the specified database file names
	for _, file := range args.GeoDB {
		db.GeoIP2, err = geoip2.Open(args.Path + file)
		if err == nil {
			break
		}
	}
	if db.GeoIP2 == nil {
		fmt.Println("ERROR: GeoIP2 database not found!")
		os.Exit(1)
	}
	defer db.GeoIP2.Close()

	// Trying to open an optional ASN base file
	db.ASN, err = geoip2.Open(args.Path + args.ASN)
	if err == nil {
		defer db.ASN.Close()
	}

	var ip net.IP
	// Try resolve domain name to IP address
	ips, err := net.LookupIP(args.Target)
	if err == nil && len(ips) > 0 {
		ip = ips[0]
	} else {
		ip = net.ParseIP(args.Target)
	}

	// Lookup GeoIP
	record, err := db.GeoIP2.City(ip)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	ti := table.NewWriter()
	ti.SetStyle(table.StyleLight)
	ti.AppendHeader(table.Row{"GeoIP INFO", "GeoIP INFO"}, table.RowConfig{AutoMerge: true})
	ti.SetOutputMirror(os.Stdout)

	var tableData []table.Row
	var reverseDNS string
	reverseAddr, err := net.LookupAddr(args.Target)
	if err == nil && len(reverseAddr) > 0 {
		reverseDNS = fmt.Sprintf(" <%s>", strings.TrimSuffix(reverseAddr[0], "."))
	}
	tableData = append(tableData, table.Row{"IP:", ip.String() + reverseDNS})
	if record.Traits.IsAnonymousProxy {
		tableData = append(tableData, table.Row{"Anonymous Proxy:", "yes"})
	}
	if record.Traits.IsSatelliteProvider {
		tableData = append(tableData, table.Row{"Satellite Provider:", "yes"})
	}
	if record.Continent.Names != nil {
		tableData = append(tableData, table.Row{"Continent:", record.Continent.Names[args.Lang]})
	}
	if record.Country.Names != nil {
		tableData = append(tableData, table.Row{"Country:", record.Country.Names[args.Lang]})
	}
	if record.Country.IsoCode != "" {
		tableData = append(tableData, table.Row{"ISO code:", record.Country.IsoCode})
	}
	if record.Country.IsInEuropeanUnion {
		tableData = append(tableData, table.Row{"European Union:", "yes"})
	}
	if record.City.Names != nil {
		tableData = append(tableData, table.Row{"City:", record.City.Names[args.Lang]})
	}
	if record.Postal.Code != "" {
		tableData = append(tableData, table.Row{"Postal Code:", record.Postal.Code})
	}
	if record.Location.TimeZone != "" {
		tableData = append(tableData, table.Row{"Timezone:", record.Location.TimeZone})
	}
	if record.Location.Latitude != 0 && record.Location.Longitude != 0 {
		tableData = append(tableData, table.Row{"Coordinates:", fmt.Sprintf("%f,%f", record.Location.Latitude, record.Location.Longitude)})
		tableData = append(tableData, table.Row{"Google Map URL:", fmt.Sprintf("https://www.google.com/maps/place/%f,%f", record.Location.Latitude, record.Location.Longitude)})
	}

	// Lookup ASN
	if db.ASN != nil {
		asn, err := db.ASN.ASN(ip)
		if err == nil {
			tableData = append(tableData, table.Row{"ASN Number:", fmt.Sprintf("%d", asn.AutonomousSystemNumber)})
			tableData = append(tableData, table.Row{"ASN Organization:", asn.AutonomousSystemOrganization})
		}
	}

	ti.AppendRows(tableData)
	ti.Render()

	if args.Whois {
		result, err := whois.Whois(args.Target)
		if err == nil {
			termWidth, _, err := term.GetSize(0)
			if err != nil {
				termWidth = 80
			}

			tw := table.NewWriter()
			tw.SetStyle(table.StyleLight)
			tw.AppendHeader(table.Row{"Whois"})
			tw.SetColumnConfigs([]table.ColumnConfig{
				{
					Number:      1,
					AlignHeader: text.AlignCenter,
					Align:       text.AlignLeft,
				},
			})
			tw.SetOutputMirror(os.Stdout)
			tw.SetAllowedRowLength(termWidth)
			tw.AppendRows([]table.Row{{result}})
			tw.Render()
		} else {
			fmt.Println("Whois error:", err)
		}
	}
}

package validation

import (
	"net/url"
	"regexp"
)

const (
	minDNSLabelLen = 3
	maxDNSLabelLen = 63
)

var (
	dnsRegexStringRFC1035Label = "^[a-z]([-a-z0-9]*[a-z0-9])?$"
	printableASCIIRegexString  = "^[\x20-\x7E]*$"
)

func IsWebURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https")
}

func IsDNSRFC1035Label(s string) bool {
	if len(s) < minDNSLabelLen || len(s) > maxDNSLabelLen {
		return false
	}
	dnsRegexRFC1035Label := regexp.MustCompile(dnsRegexStringRFC1035Label)
	return dnsRegexRFC1035Label.MatchString(s)
}

func IsPrintableASCII(s string) bool {
	printableASCIIRegex := regexp.MustCompile(printableASCIIRegexString)
	return printableASCIIRegex.MatchString(s)
}

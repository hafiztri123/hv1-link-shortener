package utils

import (
	"net"
	"net/url"
)

func IsValidURL(s string) bool {
	u, err := url.ParseRequestURI(s)
	if err != nil {
		return false
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}

	if u.Host == "" {
		return false
	}

	hostName, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		hostName = u.Host
	}

	ips, err := net.LookupIP(hostName)
	if err != nil {
		return false
	}

	for _, ip := range ips {
		if ip.IsPrivate() || ip.IsLoopback() {
			return false
		}
	}

	return true
}

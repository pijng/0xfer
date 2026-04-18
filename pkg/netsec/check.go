package netsec

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
)

var (
	ErrInvalidScheme = errors.New("invalid scheme: must be http or https")
	ErrPrivateIP    = errors.New("private addresses not allowed")
)

func IsPrivateURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("parse %q: %w", rawURL, err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return ErrInvalidScheme
	}

	host := u.Host
	if h, _, err := net.SplitHostPort(u.Host); err == nil {
		host = h
	}

	ip := net.ParseIP(host)
	if ip != nil {
		if IsPrivateIP(ip) {
			return ErrPrivateIP
		}

		return nil
	}

	resolved, err := net.DefaultResolver.LookupIP(context.Background(), "ip", host)
	if err != nil {
		return fmt.Errorf("lookup %q: %w", host, err)
	}

	for _, ip := range resolved {
		if IsPrivateIP(ip) {
			return ErrPrivateIP
		}
	}

	return nil
}

func IsPrivateIP(ip net.IP) bool {
	type cidr struct {
		mask []byte
		subnet []byte
	}

	private := []cidr{
		{mask: []byte{255, 0, 0, 0}, subnet: []byte{10, 0, 0, 0}},
		{mask: []byte{255, 240, 0, 0}, subnet: []byte{172, 16, 0, 0}},
		{mask: []byte{255, 255, 0, 0}, subnet: []byte{192, 168, 0, 0}},
		{mask: []byte{255, 0, 0, 0}, subnet: []byte{127, 0, 0, 0}},
		{mask: []byte{255, 255, 0, 0}, subnet: []byte{169, 254, 0, 0}},
		{mask: []byte{255, 0, 0, 0}, subnet: []byte{0, 0, 0, 0}},
		{mask: []byte{255, 192, 0, 0}, subnet: []byte{100, 64, 0, 0}},
		{mask: []byte{255, 255, 255, 0}, subnet: []byte{192, 0, 0, 0}},
		{mask: []byte{255, 255, 255, 0}, subnet: []byte{192, 88, 99, 0}},
		{mask: []byte{255, 254, 0, 0}, subnet: []byte{198, 18, 0, 0}},
		{mask: []byte{255, 255, 255, 0}, subnet: []byte{198, 51, 100, 0}},
		{mask: []byte{255, 255, 255, 0}, subnet: []byte{203, 0, 113, 0}},
	}

	ip4 := ip.To4()
	if ip4 == nil {
		return ip.IsLinkLocalUnicast() || ip.IsUnspecified() || ip.IsLoopback()
	}

	for _, c := range private {
		match := true
		for i := 0; i < 4; i++ {
			if ip4[i]&c.mask[i] != c.subnet[i]&c.mask[i] {
				match = false

				break
			}
		}
		if match {
			return true
		}
	}

	return ip.IsLinkLocalUnicast() || ip.IsUnspecified() || ip.IsLoopback()
}
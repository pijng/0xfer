package netsec

import (
	"net"
	"testing"
)

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"10.0.0.1", "10.0.0.1", true},
		{"10.255.255.255", "10.255.255.255", true},
		{"10.0.0.0", "10.0.0.0", true},
		{"172.16.0.0", "172.16.0.0", true},
		{"172.16.0.1", "172.16.0.1", true},
		{"172.31.255.255", "172.31.255.255", true},
		{"192.168.0.0", "192.168.0.0", true},
		{"192.168.1.1", "192.168.1.1", true},
		{"192.168.255.255", "192.168.255.255", true},
		{"127.0.0.0", "127.0.0.0", true},
		{"127.0.0.1", "127.0.0.1", true},
		{"127.255.255.255", "127.255.255.255", true},
		{"169.254.0.0", "169.254.0.0", true},
		{"169.254.0.1", "169.254.0.1", true},
		{"169.254.255.255", "169.254.255.255", true},
		{"0.0.0.0", "0.0.0.0", true},
		{"0.255.255.255", "0.255.255.255", true},
		{"100.64.0.0", "100.64.0.0", true},
		{"100.127.255.255", "100.127.255.255", true},
		{"192.0.0.0", "192.0.0.0", true},
		{"192.0.0.255", "192.0.0.255", true},
		{"198.18.0.0", "198.18.0.0", true},
		{"198.19.255.255", "198.19.255.255", true},
		{"192.88.99.0", "192.88.99.0", true},
		{"192.88.99.255", "192.88.99.255", true},
		{"198.51.100.0", "198.51.100.0", true},
		{"198.51.100.255", "198.51.100.255", true},
		{"203.0.113.0", "203.0.113.0", true},
		{"203.0.113.255", "203.0.113.255", true},
		{"public google dns", "8.8.8.8", false},
		{"public cloudflare", "1.1.1.1", false},
		{"public amazon", "3.3.3.3", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.input)
			if ip == nil {
				t.Fatalf("invalid IP: %s", tt.input)
			}
			got := IsPrivateIP(ip)
			if got != tt.want {
				t.Errorf("IsPrivateIP(%s) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsPrivateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid https", "https://google.com", false},
		{"valid http", "http://example.com", false},
		{"private ip direct", "http://192.168.1.1", true},
		{"private 10.x", "http://10.0.0.1", true},
		{"private 172.16", "http://172.16.0.1", true},
		{"private 127", "http://127.0.0.1", true},
		{"invalid scheme", "ftp://example.com", true},
		{"no scheme", "example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IsPrivateURL(tt.url)
			gotErr := err != nil
			if gotErr != tt.wantErr {
				t.Errorf("IsPrivateURL(%s) error = %v, want error = %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

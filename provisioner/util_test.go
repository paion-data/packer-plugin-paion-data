package provisioner

import (
	"testing"
)

func TestSkipConfigSSL(t *testing.T) {
	cases := []struct {
		name             string
		sslCertSource    string
		sslCertKeySource string
		domain           string
		expectedSkip     bool
		expectedErr      bool
	}{
		{"all set", "cert.pem", "key.pem", "example.com", false, false},
		{"none set", "", "", "", true, false},
		{"partial set 1", "cert.pem", "", "example.com", false, true},
		{"partial set 2", "", "key.pem", "example.com", false, true},
		{"partial set 3", "cert.pem", "key.pem", "", false, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			skip, err := SkipConfigSSL(tc.sslCertSource, tc.sslCertKeySource, tc.domain)
			if skip != tc.expectedSkip {
				t.Errorf("expected %v, got %v", tc.expectedSkip, skip)
			}
			if (err != nil) != tc.expectedErr {
				t.Errorf("expected error %v, got %v", tc.expectedErr, err != nil)
			}
		})
	}
}

func TestGetHomeDir(t *testing.T) {
	cases := []struct {
		name         string
		configValue  string
		expectedHome string
	}{
		{"empty input", "", "/root"},
		{"non-empty input", "/custom/home/dir", "/custom/home/dir"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetHomeDir(tc.configValue)
			if result != tc.expectedHome {
				t.Errorf("expected %q, got %q", tc.expectedHome, result)
			}
		})
	}
}

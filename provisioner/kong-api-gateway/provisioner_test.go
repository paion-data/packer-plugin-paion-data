// Copyright (c) Jiaqi Liu
// SPDX-License-Identifier: MPL-2.0

package kongApiGateway

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getHomeDir(t *testing.T) {
	data := []struct {
		name        string
		configValue string
		expected    string
	}{
		{"regular directory is specified", "/", "/"},
		{"no directory is specified as home dir", "", "/home/ubuntu"},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			actual := getHomeDir(d.configValue)
			if actual != d.expected {
				t.Errorf("Expected %s, got %s", d.expected, actual)
			}
		})
	}
}

func Test_skipConfigSSL(t *testing.T) {
	tests := []struct {
		name          string
		config        Config
		expectedSkip  bool
		expectedError error
	}{
		{
			name: "All configurations set",
			config: Config{
				SslCertSource:        "cert.pem",
				SslCertKeySource:     "key.pem",
				KongApiGatewayDomain: "example.com",
			},
			expectedSkip:  false,
			expectedError: nil,
		},
		{
			name: "No configurations set",
			config: Config{
				SslCertSource:        "",
				SslCertKeySource:     "",
				KongApiGatewayDomain: "",
			},
			expectedSkip:  true,
			expectedError: nil,
		},
		{
			name: "Partial configurations set",
			config: Config{
				SslCertSource:        "cert.pem",
				SslCertKeySource:     "",
				KongApiGatewayDomain: "",
			},
			expectedSkip:  false,
			expectedError: fmt.Errorf("sslCertSource, sslCertKeySource and kongApiGatewayDomain must be set together"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			provisioner := &Provisioner{config: tc.config}
			skip, err := provisioner.skipConfigSSL()
			assert.Equal(t, tc.expectedSkip, skip, "SkipSSL should be %v", tc.expectedSkip)
			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.expectedError.Error(), "Expected error message: %s", tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

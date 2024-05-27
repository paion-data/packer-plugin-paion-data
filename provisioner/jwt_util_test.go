package provisioner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCoordsFromJWKS(t *testing.T) {
	cases := []struct {
		name      string
		base64X   string
		base64Y   string
		expectedX string
		expectedY string
	}{
		{"JWKOfTrue", "WFhY", "WVlZ", "XXX", "YYY"},
		{"JWKOfNil", "", "", "", ""},
		{"JWKOfErr", "WA=", "WQ=", "", ""},
		{"JWKSNoKeys", "", "", "", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "JWKSNoKeys" {
				testJWKS := JWKS{}
				x, y, err := getCoordsFromJWKS(testJWKS)
				assert.Error(t, err)
				assert.Nil(t, x)
				assert.Nil(t, y)
				return
			}

			testJWK := JWK{
				X: tc.base64X,
				Y: tc.base64Y,
			}
			testJWKS := JWKS{
				Keys: []JWK{testJWK},
			}

			x, y, err := getCoordsFromJWKS(testJWKS)

			if tc.name == "JWKOfErr" {
				assert.Error(t, err)
				assert.Nil(t, x)
				assert.Nil(t, y)
				return
			}

			assert.NoError(t, err)

			if string(x) != tc.expectedX {
				t.Errorf("expected %v, got %v", tc.expectedX, x)
			}
			if string(y) != tc.expectedY {
				t.Errorf("expected %v, got %v", tc.expectedY, y)
			}

			assert.NotNil(t, x)
			assert.NotNil(t, y)
			assert.Equal(t, string(x), tc.expectedX)
			assert.Equal(t, string(y), tc.expectedY)
		})
	}
}

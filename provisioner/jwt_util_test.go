package provisioner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCoordsFromJWKS(t *testing.T) {
	JWKSOfNoKeys := JWKS{}
	x, y, err := getCoordsFromJWKS(JWKSOfNoKeys)
	assert.Error(t, err)
	assert.Nil(t, x)
	assert.Nil(t, y)

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
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testJWK := JWK{
				X: tc.base64X,
				Y: tc.base64Y,
			}
			testJWKS := JWKS{
				Keys: []JWK{testJWK},
			}

			x, y, _ := getCoordsFromJWKS(testJWKS)

			if string(x) != tc.expectedX {
				t.Errorf("expected %v, got %v", tc.expectedX, x)
			}
			if string(y) != tc.expectedY {
				t.Errorf("expected %v, got %v", tc.expectedY, y)
			}
		})
	}
}

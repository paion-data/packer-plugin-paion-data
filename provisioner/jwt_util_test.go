package provisioner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCoordsFromJWKS(t *testing.T) {
	testJWK := JWK {
		Kty: "test",
		Use: "test",
		Kid: "test",
		Alg: "test",
		Crv: "test",
		X  : "WFhY",
		Y  : "WVlZ",
	}

	testJWKS := JWKS {
		Keys: []JWK{testJWK},
	}

	x, y, err := getCoordsFromJWKS(testJWKS)

	assert.NoError(t, err)
	
	assert.NotNil(t, x)
	assert.NotNil(t, y)

	assert.Equal(t, string(x), "XXX")
	assert.Equal(t, string(y), "YYY")
}

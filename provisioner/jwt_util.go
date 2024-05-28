package provisioner

import (
    "encoding/json"
    "encoding/base64"
    "encoding/pem"
    "fmt"
    "io/ioutil"
    "net/http"
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/x509"
    "math/big"
)

type JWKS struct {
    Keys []JWK `json:"keys"`
}

type JWK struct {
    Kty string `json:"kty"`
    Use string `json:"use"`
    Kid string `json:"kid"`
    Alg string `json:"alg"`
    Crv string `json:"crv"`
    X   string `json:"x"`
    Y   string `json:"y"`
}

func GetJWKSPublicKeyPEM(url string) (string, error) {
    httpClient := &http.Client{}
    resp, err := httpClient.Get(url)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    // check response status code
    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("error fetching JWKS: %s", resp.Status)
    }

    // load response body
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    // get JWKS from body
    var jwks JWKS
    err = json.Unmarshal(body, &jwks)
    if err != nil {
        return "", err
    }
    
    x, y, err := getCoordsFromJWKS(jwks)
    if err != nil {
        return "", err
    }

    // Create the corresponding elliptic curve according to crv
    curve := elliptic.P384()

    // create ECDSA public key
    xInt := new(big.Int).SetBytes(x)
    yInt := new(big.Int).SetBytes(y)

    ecPubKey := &ecdsa.PublicKey{
        Curve: curve,
        X:     xInt,
        Y:     yInt,
    }

    // Encode the ECDSA public key in PEM format
    pemPubKey, err := pemEncodeECDSAPublicKey(ecPubKey)
    if err != nil {
        return "", err
    }
    
    return string(pemPubKey), nil
}

func pemEncodeECDSAPublicKey(key *ecdsa.PublicKey) ([]byte, error) {
    // Encode the ECDSA public key in DER format
    derBytes, err := x509.MarshalPKIXPublicKey(key)
    if err != nil {
        return nil, err
    }
    
    // Converts a DER encoded public key to PEM format
    pemBlock := &pem.Block{
        Type: "PUBLIC KEY",
        Bytes: derBytes,
    }
    return pem.EncodeToMemory(pemBlock), nil
}

func getCoordsFromJWKS(jwks JWKS) ([]byte, []byte, error) {
    if len(jwks.Keys) == 0 {
        return nil, nil, fmt.Errorf("no keys found in JWKS")
    }

    // JWKS which we need has only one JWK
    jwk := jwks.Keys[0]

    // Decode X and Y
    x, err := base64.RawURLEncoding.DecodeString(jwk.X)
    if err != nil {
        return nil, nil, err
    }
    y, err := base64.RawURLEncoding.DecodeString(jwk.Y)
    if err != nil {
        return nil, nil, err
    }

    return x, y, nil
}

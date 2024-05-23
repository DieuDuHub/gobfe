package main

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func JWTMiddleware(jwkSet map[string]*rsa.PublicKey) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// Extract the key ID (kid) from the token header
			kid, ok := token.Header["kid"].(string)
			if !ok {
				return nil, fmt.Errorf("kid not found in token header")
			}

			// Retrieve the public key from the JWK set using the key ID
			publicKey, keyExists := jwkSet[kid]
			if !keyExists {
				return nil, fmt.Errorf("public key not found for kid: %v", kid)
			}

			return publicKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Extract the mainId claim from the token
		/* mainIDStr, ok := token.Claims.(jwt.MapClaims)["mainId"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "mainId not found in token"})
			c.Abort()
			return
		}

		// Convert the mainId string to an integer
				mainID, err := strconv.Atoi(mainIDStr)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Error converting mainId to integer"})
					c.Abort()
					return
				}

				// Store the mainId in the request context
				c.Set("mainId", mainID)
		*/
		// Continue with the next middleware or route handler
		c.Next()
	}
}

// GetJWKSet retrieves the JWK set from the specified URL and returns a map of RSA public keys.
func GetJWKSet(url string) (map[string]*rsa.PublicKey, error) {
	// Make the GET request
	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Error making GET request: %v", err)
	}
	defer response.Body.Close()

	// Decode the JSON response
	var jwkSet struct {
		Keys []struct {
			Kid string   `json:"kid"`
			N   string   `json:"n"`
			E   string   `json:"e"`
			X5C []string `json:"x5c"`
		} `json:"keys"`
	}
	decoder := json.NewDecoder(response.Body)
	if err := decoder.Decode(&jwkSet); err != nil {
		return nil, fmt.Errorf("Error decoding JSON: %v", err)
	}

	// Create a map to store RSA public keys
	jwkMap := make(map[string]*rsa.PublicKey)

	// Iterate through each key in the JWK set
	for _, key := range jwkSet.Keys {
		// Decode base64url-encoded modulus (N) and exponent (E)
		modulus, err := decodeBase64URL(key.N)
		if err != nil {
			return nil, fmt.Errorf("Error decoding modulus: %v", err)
		}

		exponent, err := decodeBase64URL(key.E)
		if err != nil {
			return nil, fmt.Errorf("Error decoding exponent: %v", err)
		}

		// Create RSA public key
		pubKey := &rsa.PublicKey{
			N: modulus,
			E: int(exponent.Int64()),
		}

		// Store the public key in the map using the key ID (Kid)
		jwkMap[key.Kid] = pubKey
	}

	return jwkMap, nil
}

// decodeBase64URL decodes a base64url-encoded string and returns a big.Int
func decodeBase64URL(input string) (*big.Int, error) {
	// Convert base64url to base64
	base64Str := strings.ReplaceAll(input, "-", "+")
	base64Str = strings.ReplaceAll(base64Str, "_", "/")

	// Pad the base64 string with "="
	switch len(base64Str) % 4 {
	case 2:
		base64Str += "=="
	case 3:
		base64Str += "="
	}

	// Decode base64 string
	data, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, err
	}

	// Convert bytes to big.Int
	result := new(big.Int).SetBytes(data)
	return result, nil
}

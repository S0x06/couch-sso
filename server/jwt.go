package server

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

func PublicKey() (file []byte) {
	file, e := ioutil.ReadFile("./public.pem")
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}

	return file
}

// https://tools.ietf.org/html/rfc7519#section-4.1
// https://jwt.io/introduction/
// DecodeJWT decodes a JWT and returns the payload as a map[string]interface{}.
func DecodeJWT(jwt string) (map[string]interface{}, error) {
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return nil, errors.New("Invalid JWT Structure. ")
	}

	header, _ := base64.RawStdEncoding.DecodeString(parts[0])
	payload, _ := base64.RawStdEncoding.DecodeString(parts[1])
	//signature, _ := base64.StdEncoding.DecodeString(parts[2])

	// JSON decode header
	var headdat map[string]string
	if err := json.Unmarshal(header, &headdat); err != nil {
		log.Println(err)
	}

	// JSON decode payload
	var pldat map[string]interface{}
	if err := json.Unmarshal(payload, &pldat); err != nil {
		log.Println(err)
	}

	expTime := time.Unix(int64(pldat["exp"].(float64)), 9).Local() // 2018-06-12 21:06:28.000000009 +0800 CST
	now := time.Now().Local()                                      //2018-06-06 13:31:45.451336189 +0800 CST

	if now.After(expTime) {
		return nil, errors.New("Expired JWT. ")
	}

	return pldat, nil
}

func VerifyToken(jwt string) (err error) {
	block, _ := pem.Decode(PublicKey())
	pubInterface, _ := x509.ParsePKIXPublicKey(block.Bytes)

	pub := pubInterface.(*rsa.PublicKey)

	err = Verify(jwt, pub)
	return err
}

// Verify tests whether the provided JWT token's signature was produced by the private key
// associated with the supplied public key.
func Verify(token string, key *rsa.PublicKey) error {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return errors.New("jws: invalid token received, token must have 3 parts")
	}

	signedContent := parts[0] + "." + parts[1]
	signatureString, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return err
	}

	h := sha256.New()
	h.Write([]byte(signedContent))
	return rsa.VerifyPKCS1v15(key, crypto.SHA256, h.Sum(nil), []byte(signatureString))
}

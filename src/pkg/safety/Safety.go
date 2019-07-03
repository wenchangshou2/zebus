package safety

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

type Safety struct {
	privateKey string
	publicKey  string
}

func (s *Safety) DefaultKey() {
	var private_key string = `MIICWwIBAAKBgQCqsRHp2WYgrkI1qrS+7T/vWvI4z4sUNwKVsXnMaDnY60c11923
nRM/FHyYAF5qF6KlbQ0aqf8DBz8bDTsyUHb+jrNu+SAoQjVO26AIvAVhogK5qLUz
6D6xomngyo93zjH1wO4ptc8x02Mlumju6YLKNWpKIR5/MRj7vYz8FtZ8bwIDAQAB
AoGALfLgqZzWOzHtrNi5MzRWo65NyjFEdTqhvX47FWVxPQ2I69uiWc004yQ2rgxb
Xh/irrl+b5EXjs8ik7uqFc9HWKqV1O9kNpAS6qla1yxUPLCIBGpxcErrk6GnPpxp
eU4Se5X41n1A/bGtINks29n6YhAVxdiUMFMVlGp9ARjesmECQQDaf5RX5Ne+w/vM
S9Bo7GzZuf81aTgILur1a5U7vXnABeTXiAODALhgcwoMmT7JN/8dRhAwk0XRImot
m33zfcGpAkEAx/zztz4SfXFXwegMdLaO2N/NIjZhj9kBFBg1KH0bsWwI5Sfcr27d
Azi94GZ4N+IkAoXTv9DkWhooCd8oNO/MVwJAHaK6QyWl4ZkBeRc7YE/Y/7sLk3n/
AJUkhz8dUaoEbngeLuGi4EzjtSlFTqomavJuZtEO9xeym4gYcLErZzBCaQJAN6Hn
TkdHL3wzNG7P4DvUmwIO94B3PWPZh/R//SZoaL+r7ctb+bV2Z+oF8AGxWaJf8A+4
avi6PVJfZvecILXAewJAIVfQIZPH3BmKqwfPNn9Y7J8+o5Uc6b4Brk/5VNyWHcWK
sOgeRICIbqubBO3vXmNeaJPDV5B28sVnSTgWqf0Wdg==
`
	var public_key string = `
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCqsRHp2WYgrkI1qrS+7T/vWvI4
z4sUNwKVsXnMaDnY60c11923nRM/FHyYAF5qF6KlbQ0aqf8DBz8bDTsyUHb+jrNu
+SAoQjVO26AIvAVhogK5qLUz6D6xomngyo93zjH1wO4ptc8x02Mlumju6YLKNWpK
IR5/MRj7vYz8FtZ8bwIDAQAB
`
	s.privateKey = private_key
	s.publicKey = public_key
}
func (s *Safety) SignWithSha1Hex(data string, prvKey string) (string, error) {
	keyByts, err := hex.DecodeString(prvKey)
	if err != nil {
		return "", err
	}
	privateKey, err := x509.ParsePKCS8PrivateKey(keyByts)
	if err != nil {
		fmt.Println("ParsePKCS8PrivateKey err", err)
		return "", err
	}
	h := sha1.New()
	h.Write([]byte([]byte(data)))
	hash := h.Sum(nil)
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey.(*rsa.PrivateKey), crypto.SHA1, hash[:])
	if err != nil {
		fmt.Printf("Error from signing: %s\n", err)
		return "", err
	}
	out := hex.EncodeToString(signature)
	return out, nil
}
func (s *Safety) EncryptWithSha1Base64(originalData string) (string, error) {
	key, err := base64.StdEncoding.DecodeString(s.publicKey)
	pubKey, _ := x509.ParsePKIXPublicKey(key)
	encryptedData, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey.(*rsa.PublicKey), []byte(originalData))
	return base64.StdEncoding.EncodeToString(encryptedData), err
}
func (s *Safety) DecryptWithSha1Base64(encryptedData string) (string, error) {
	encryptedDecodeBytes, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", err
	}
	key, _ := base64.StdEncoding.DecodeString(s.privateKey)
	prvKey, _ := x509.ParsePKCS1PrivateKey(key)
	originalData, err := rsa.DecryptPKCS1v15(rand.Reader, prvKey, encryptedDecodeBytes)
	return string(originalData), err
}
func (s *Safety) VerySignWithSha1Base64(originalData, signData, pubKey string) error {
	sign, err := base64.StdEncoding.DecodeString(signData)
	if err != nil {
		return err
	}
	public, _ := base64.StdEncoding.DecodeString(pubKey)
	pub, err := x509.ParsePKIXPublicKey(public)
	if err != nil {
		return err
	}
	hash := sha1.New()
	hash.Write([]byte(originalData))
	return rsa.VerifyPKCS1v15(pub.(*rsa.PublicKey), crypto.SHA1, hash.Sum(nil), sign)
}

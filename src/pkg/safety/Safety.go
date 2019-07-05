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
var (
	G_Safety *Safety=&Safety{}

)
func init(){
	G_Safety.DefaultKey()
}

func (s *Safety) DefaultKey() {
	var private_key string = `
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
reBNAoIBABcv1O+7h4MnWvtb72gU+o8FUDM0EPtVw2xILW9RVgVy70V2WubLuBkz
U4iud6GSyLUti8QTeQ8rkqjL7vpFXAMj/g7zQs+F9m647NF7ojdXn2fjHTFt0M3I
RLK0K/pKk6IK2fQDOfNhKZcyKFRsqEzAiLnbF1CzivkosRyUlmBEd1P53mKUSLaH
vhA8eWhoR6m1/u3KFcQ4Q1xNsB3Cm2QHPqQLJ7IhZloMpcRA4lcsrquuvrDHcUt3
FYQkQaxL3fi10iNzmPiHb/Ax0XpfUZA/RYeli4B6nD3LALMYXpPQxDF8XeJrBGAY
zx59W57VzvYo3lcAQhJN+1LN3sB8CIECggEBAMaQ9a64SCyhnsi0VT5LAomsZSfU
Woekb2SLcP3OWk1th8wyKcFA06Gf802gLsm9VXRZ/CTiCnnsRsyj70yHmcfmK78z
Xhnb3ySUut1ChFZm09YPITKkIlcUrzbHJPaO+97mZWhdWCYa6p6BjOcTKA9GFKhw
TzDlGBkwdaUE+aVsNBLflJT7zEfOTB3uqqRIwzyBsOQltaU5dXztiA0Iw9xMnWc0
zv3Fo2YHmhzNu73iV3SnnzyySzinQWpKjnxRMvLHcZ3EaFwHuDpMtIxNrBOojXZt
zJoKNmatUYk+b8Epc/sQk2dZ+z3tylztvX1Ohx0bnXq794SAKp1x94LM/ZU=
`
	var public_key string = `
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAyofRhbjcABvv5mUY1Em/
bMf6XjPZvonfdslzRbpJE8nArUY5QMnQ53OOc1I9syYmSVp/qAMms5x7gdhrNlmn
HCn84aZmzmTTG0BzAwejP8F+oTSNGwaFfA3PvF7jbyb1UOxr54m1uqfIeG60ruvO
0L8BscEhfhiWAZZ4l0BGdt8lcjC1yqiLKxatSYU+v2GLBs07Yx5BOLCEZ0cnvEUk
ROJvedHtRE4huTYjz5EgkeIEeMmzk/nkDJF0qidb/yVGU2QogqsswMQ6S/BdANQE
+MhfIMRco9NnfiTXjMJtzo7wkkdWxyWB/W7dyhi7JLmDKT76B5mahLVBFdX+fQsW
PynrZyHf0WVlbFLLDwYE3vxfDq9D+7fYQwHqPYCyI6ZT55H696uv0v/P9GUSci5L
/AmtiwowYErOl/lyZSNUP9g59WHN30j4eHU3VWuX0d3A4F/zINFQfcWdfOEqE5kU
BILhBn1a15x2XIWd1vYz7lHxfyWM2Ns7KWpNxnOGvxucunIPyVBlBXli7wPbfwKx
cTC2qZseBocvEhew8T7aDNVMWvn3C9LeINbqtII+vHvkwUfP3hf3Z5l93Xi0vJ0H
7+hLEJVOliqwFVJNylY+gYiR8TVFjjW/NlcU2U2tpkL4DTwVykd3Ob1MOdbDKG4P
8oXek3cSBMRINRt4zcY74p0CAwEAAQ==
`
	s.privateKey = private_key
	s.publicKey = public_key
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

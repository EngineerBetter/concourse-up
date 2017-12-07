package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"bytes"

	"golang.org/x/crypto/ssh"
)

// GenerateSSHKeyPair generates a new ssh public key pair
// http://stackoverflow.com/questions/21151714/go-generate-an-ssh-public-key
func GenerateSSHKeyPair() ([]byte, []byte, string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, "", err
	}

	privateKeyBuffer := bytes.NewBuffer(nil)

	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	if err = pem.Encode(privateKeyBuffer, privateKeyPEM); err != nil {
		return nil, nil, "", err
	}

	// generate and write public key
	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, "", err
	}
	pubKeyBytes := ssh.MarshalAuthorizedKey(pub)

	// generate public key’s MD5 fingerprint
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(pubKeyBytes)
	if err != nil {
		return nil, nil, "", err
	}

	fingerprint := ssh.FingerprintLegacyMD5(pubKey)

	return privateKeyBuffer.Bytes(), pubKeyBytes, fingerprint, nil
}

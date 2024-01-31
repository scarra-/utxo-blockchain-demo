package wallet

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"io"
)

const (
	checksumLength = 4
	version        = byte(0x00)
)

type Wallet struct {
	PrivateKey ed25519.PrivateKey
}

func MakeWallet() *Wallet {
	private := NewKeyPair()
	return &Wallet{
		PrivateKey: private,
	}
}

func (w Wallet) Address() []byte {
	pubHash := PublicKeyHash([]byte(w.PrivateKey.Public().(ed25519.PublicKey)))
	versionedHash := append([]byte{version}, pubHash...)
	checksum := Checksum(versionedHash)

	fullHash := append(versionedHash, checksum...)
	address := Base58Encode(fullHash)

	return address
}

func (w Wallet) PublicKeyBytes() []byte {
	return []byte(w.PrivateKey.Public().(ed25519.PublicKey))
}

func ValidateAddress(address string) bool {
	pubKeyHash := Base58Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-checksumLength:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-checksumLength]
	targetChecksum := Checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(actualChecksum, targetChecksum) == 0
}

func NewKeyPair() ed25519.PrivateKey {
	seed := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, seed)
	if err != nil {
		panic(err)
	}

	return ed25519.NewKeyFromSeed(seed)
}

func PublicKeyHash(pubKey []byte) []byte {
	hash := sha256.Sum256(pubKey)
	return hash[:]
}

func Checksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])

	return secondHash[:checksumLength]
}

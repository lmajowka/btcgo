package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"golang.org/x/crypto/ripemd160"
)

func GenerateWif(privKeyInt *big.Int) string {
	privKeyHex := fmt.Sprintf("%064x", privKeyInt)

	// Decode the hexadecimal private key
	privKeyBytes, err := hex.DecodeString(privKeyHex)
	if err != nil {
		log.Fatal(err)
	}

	// Add prefix and sufix
	extendedKey := append([]byte{byte(0x80)}, privKeyBytes...)
	extendedKey = append(extendedKey, byte(0x01))

	// Calc checksum
	firstSHA := sha256.Sum256(extendedKey)
	secondSHA := sha256.Sum256(firstSHA[:])
	checksum := secondSHA[:4]

	// Add checksum
	finalKey := append(extendedKey, checksum...)

	// Encode to base58
	wif := Encode(finalKey)

	return wif
}

func CreatePublicHash160(privKeyInt *big.Int) []byte {

	privKeyBytes := privKeyInt.Bytes()

	// Create a new private key using the secp256k1 package
	privKey := secp256k1.PrivKeyFromBytes(privKeyBytes)

	// Get the corresponding public key in compressed format
	compressedPubKey := privKey.PubKey().SerializeCompressed()

	// Generate a Bitcoin address from the public key
	pubKeyHash := hash160(compressedPubKey)

	return pubKeyHash

}

// hash160 computes the RIPEMD160(SHA256(b)) hash.
func hash160(b []byte) []byte {
	h := sha256.New()
	h.Write(b)
	sha256Hash := h.Sum(nil)

	r := ripemd160.New()
	r.Write(sha256Hash)
	return r.Sum(nil)
}

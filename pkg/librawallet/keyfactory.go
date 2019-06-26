package librawallet

import (
	"encoding/binary"
	"io"
	"log"

	"github.com/codemaveric/libra-go/pkg/types"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/sha3"
)

const (
	MNEMONIC_SALT_PREFIX string = "LIBRA WALLET: mnemonic salt prefix$"
	MASTER_KEY_SALT      string = "LIBRA WALLET: master key salt$"
	INFO_PREFIX          string = "LIBRA WALLET: derived key$"
)

const (
	PublicKeySize = 32
)

type ExtendedPrivKey struct {
	ChildNumber uint64
	PrivateKey  ed25519.PrivateKey
}

func (e *ExtendedPrivKey) GetAddress() types.AccountAddress {
	publicKey := e.GetPublic()
	// keccaksha3 := sha3.NewLegacyKeccak256()
	var accountAddress types.AccountAddress
	digest := sha3.Sum256(publicKey)
	accountAddress = digest[:]
	return accountAddress
}

func (e *ExtendedPrivKey) GetPublic() ed25519.PublicKey {
	publicKey := make([]byte, PublicKeySize)
	copy(publicKey, e.PrivateKey[32:])
	return publicKey
}

type KeyFactory struct {
	Master []byte
	Seed   *Seed
}

func NewKeyFactory(seed *Seed) *KeyFactory {
	extract := hkdf.Extract(sha3.New256, seed.Data, []byte(MASTER_KEY_SALT))
	return &KeyFactory{Master: extract, Seed: seed}
}

func (k *KeyFactory) GenerateKey(childNumber uint64) *ExtendedPrivKey {
	numByte := make([]byte, 8)
	binary.LittleEndian.PutUint64(numByte, childNumber)
	info := append([]byte(INFO_PREFIX), numByte...)
	r := hkdf.Expand(sha3.New256, k.Master, info)

	seed := make([]byte, 32)
	n, _ := io.ReadFull(r, seed)
	log.Println(seed)
	log.Println(n)
	privateKey := ed25519.NewKeyFromSeed(seed)
	return &ExtendedPrivKey{ChildNumber: childNumber, PrivateKey: privateKey}
}

type Seed struct {
	Data []byte
}

/// This constructor implements the one-way function that allows to generate a Seed from a
/// particular Mnemonic and salt. WalletLibrary implements a fixed salt, but a user could
/// choose a user-defined salt instead of the hardcoded one.
func NewSeed(mnemonic Mnemonic, salt string) *Seed {
	if salt == "" {
		salt = "LIBRA"
	}
	mnemonicsalt := []byte(MNEMONIC_SALT_PREFIX + salt)
	data := pbkdf2.Key(mnemonic.ToBytes(), mnemonicsalt, 2048, 32, sha3.New256)
	return &Seed{data}
}
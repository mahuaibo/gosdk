package bvm

import (
	"crypto/rand"
	"encoding/pem"
	"errors"
	"io/ioutil"

	"github.com/hyperchain/gosdk/common"
	gm "github.com/ultramesh/crypto-gm"
	"github.com/ultramesh/crypto-standard/asym"
	"github.com/ultramesh/crypto-standard/hash"
	"github.com/ultramesh/flato-msp-cert/primitives"
)

// KeyPair privateKey(ecdsa.PrivateKey or guomi.PrivateKey) and publicKey string
type KeyPair struct {
	privKey interface{}
	pubKey  string
}

//ParsePriv parse key pair by file path
func ParsePriv(k []byte) (*KeyPair, error) {
	var key []byte
	block, _ := pem.Decode(k)
	if block != nil {
		key = block.Bytes
	}

	newKey, err := primitives.UnmarshalPrivateKey(key)
	if err != nil {
		return nil, err
	}
	var pub []byte
	switch key := newKey.(type) {
	case *asym.ECDSAPrivateKey:
		pub, _ = primitives.MarshalPublicKey(key.Public())
	case *gm.SM2PrivateKey:
		pub, _ = primitives.MarshalPublicKey(key.Public())
	}
	keyPair := &KeyPair{
		privKey: newKey,
		pubKey:  common.Bytes2Hex(pub),
	}
	return keyPair, nil
}

// Sign sign the message by privateKey
func (key *KeyPair) Sign(msg []byte) ([]byte, error) {
	switch key.privKey.(type) {
	case *asym.ECDSAPrivateKey:
		//to maintain compatibility, sdkcert's signature is always sha256
		h, _ := hash.NewHasher(hash.SHA2_256).Hash(msg)
		data, err := key.privKey.(*asym.ECDSAPrivateKey).Sign(rand.Reader, h, nil)
		if err != nil {
			return nil, err
		}
		return data, nil
	case *gm.SM2PrivateKey:
		gmKey := key.privKey.(*gm.SM2PrivateKey)
		h := gm.HashBeforeSM2(gmKey.Public().(*gm.SM2PublicKey), msg)
		data, err := gmKey.Sign(rand.Reader, h, nil)
		if err != nil {
			return nil, err
		}
		return data, nil
	default:
		common.GetLogger("rpc").Error("unsupported sign type")
		return nil, errors.New("signature type error")
	}
}

// NewKeyPair create a new KeyPair(ecdsa or sm2)
func NewKeyPair(privFilePath string) (*KeyPair, error) {
	k, err := ioutil.ReadFile(privFilePath)
	if err != nil {
		return nil, err
	}
	return ParsePriv(k)
}

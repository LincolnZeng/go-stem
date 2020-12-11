/**
*  @file
*  @copyright defined in scdo/LICENSE
 */

package common

import (
	"bytes"
	"errors"
	"math/big"
	"strings"

	"github.com/scdoproject/go-scdo/common/hexutil"
)

const (
	// HashLength the leghth of hash
	HashLength = 32
)

// EmptyHash empty hash
var EmptyHash = Hash{}

// ErrOnly0xPrefix the string is invalid length only 0x or 0X prefix
var ErrOnly0xPrefix = errors.New("the string is invalid length only 0x or 0X prefix")

// Hash the hash value generated by sha-3
type Hash [HashLength]byte

// BytesToHash converts bytes to hash value
func BytesToHash(b []byte) Hash {
	a := &Hash{}
	a.SetBytes(b)
	return *a
}

// StringToHash converts a string to the hash
func StringToHash(s string) Hash {
	return BytesToHash([]byte(s))
}

// SetBytes sets the hash to the value of b.
func (a *Hash) SetBytes(b []byte) {
	if len(b) > HashLength {
		b = b[len(b)-HashLength:]
	}

	copy(a[HashLength-len(b):], b)
}

// Bytes returns its actual bits
func (a Hash) Bytes() []byte {
	return a[:]
}

// String returns the string representation of the hash
func (a Hash) String() string {
	return a.Hex()
}

// Equal returns a boolean value indicating whether the hash a is equal to the input hash b.
func (a *Hash) Equal(b Hash) bool {
	return bytes.Equal(a[:], b[:])
}

// Hex returns the hex form of the hash
func (a Hash) Hex() string {
	return hexutil.BytesToHex(a[:])
}

// HexToHash return the hash form of the hex
func HexToHash(hex string) (Hash, error) {
	if strings.EqualFold(hex, "0x") {
		return EmptyHash, ErrOnly0xPrefix
	}
	byte, err := hexutil.HexToBytes(hex)
	if err != nil {
		return EmptyHash, err
	}

	hash := BytesToHash(byte)
	return hash, nil
}

func MustHexToHash(hex string) Hash {
	hash, err := HexToHash(hex)
	if err != nil {
		panic(err)
	}

	return hash
}

// IsEmpty return true if this hash is empty. Otherwise, false.
func (a Hash) IsEmpty() bool {
	return a == EmptyHash
}

// BigToHash converts a big int to Hash.
func BigToHash(b *big.Int) Hash { return BytesToHash(b.Bytes()) }

// Big converts this Hash to a big int.
func (a Hash) Big() *big.Int { return new(big.Int).SetBytes(a[:]) }

// MarshalText marshals the Hash to HEX string.
func (a Hash) MarshalText() ([]byte, error) {
	str := a.Hex()
	return []byte(str), nil
}

// UnmarshalText unmarshals Hash from HEX string.
func (a *Hash) UnmarshalText(json []byte) error {
	b, err := HexToHash(string(json))
	if err != nil {
		return err
	}

	copy(a[:], b[:])
	return nil
}

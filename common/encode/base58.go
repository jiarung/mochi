package encode

import (
	"math/big"

	"github.com/satori/go.uuid"
	"github.com/tv42/base58"
)

// UUIDToBase58 encodes an UUID(128 bits) to base58 byte slice.
func UUIDToBase58(u uuid.UUID) []byte {
	var bigInt big.Int
	bigInt.SetBytes(u.Bytes())
	return base58.EncodeBig(nil, &bigInt)
}

// Base58ToUUID decodes a base58 byte slice to an UUID(128 bits).
func Base58ToUUID(b []byte) (uuid.UUID, error) {
	bigInt, err := base58.DecodeToBig(b)
	if err != nil {
		return uuid.Nil, nil
	}
	return uuid.FromBytes(bigInt.Bytes())
}

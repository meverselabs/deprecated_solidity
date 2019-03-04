package vm

import (
	"math/big"

	"github.com/fletaio/common"
	"github.com/fletaio/common/hash"
)

// BytesToAddress get a address from the bytes
func BytesToAddress(bs []byte) common.Address {
	var addr common.Address
	if len(bs) > common.AddressSize {
		bs = bs[:common.AddressSize]
	}
	copy(addr[:], bs[:])
	return addr
}

// BytesToHash get a hash from the bytes
func BytesToHash(bs []byte) hash.Hash256 {
	var h hash.Hash256
	if len(bs) > hash.Hash256Size {
		bs = bs[:hash.Hash256Size]
	}
	copy(h[:], bs[:])
	return h
}

// AddressToBig get a big from the bytes
func AddressToBig(addr common.Address) *big.Int {
	v := new(big.Int)
	v.SetBytes(addr[:])
	return v
}

// HashToBig get a big from the hash
func HashToBig(h hash.Hash256) *big.Int {
	v := new(big.Int)
	v.SetBytes(h[:])
	return v
}

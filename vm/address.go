package vm

import (
	"math/big"

	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/common/hash"
)

// BytesToAddress TODO
func BytesToAddress(bs []byte) common.Address {
	var addr common.Address
	if len(bs) > common.AddressSize {
		bs = bs[:common.AddressSize]
	}
	copy(addr[:], bs[:])
	return addr
}

// BytesToHash TODO
func BytesToHash(bs []byte) hash.Hash256 {
	var h hash.Hash256
	if len(bs) > hash.Hash256Size {
		bs = bs[:hash.Hash256Size]
	}
	copy(h[:], bs[:])
	return h
}

// AddressToBig TODO
func AddressToBig(addr common.Address) *big.Int {
	v := new(big.Int)
	v.SetBytes(addr[:])
	return v
}

// HashToBig TODO
func HashToBig(h hash.Hash256) *big.Int {
	v := new(big.Int)
	v.SetBytes(h[:])
	return v
}

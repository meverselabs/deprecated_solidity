package solidity

import (
	"encoding/binary"
	"log"

	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/common/hash"
	"git.fleta.io/fleta/core/amount"
	"git.fleta.io/fleta/core/data"
	"git.fleta.io/fleta/solidity/vm"
)

// ViewDB errors
var ()

// ViewDB is an EVM database for full state querying.
// It doesn't allow any modification of DB
// It is used to execute view query
type ViewDB struct {
	ChainCoord *common.Coordinate
	Loader     data.Loader
}

// CreateAccount is not allowed
func (sd *ViewDB) CreateAccount(addr common.Address) {
	panic(ErrNotAllowed)
}

// SubBalance is not allowed
func (sd *ViewDB) SubBalance(addr common.Address, b *amount.Amount) {
	panic(ErrNotAllowed)
}

// AddBalance is not allowed
func (sd *ViewDB) AddBalance(addr common.Address, b *amount.Amount) {
	panic(ErrNotAllowed)
}

// GetBalance returns the target chain balance from the account of the address
func (sd *ViewDB) GetBalance(addr common.Address) *amount.Amount {
	acc, err := sd.Loader.Account(addr)
	if err != nil {
		panic(err)
	}
	return acc.Balance(sd.ChainCoord)
}

// GetSeq returns the sequence of the address
func (sd *ViewDB) GetSeq(addr common.Address) uint64 {
	return sd.Loader.Seq(addr)
}

// AddSeq is not allowed
func (sd *ViewDB) AddSeq(addr common.Address) {
	panic(ErrNotAllowed)
}

// GetCodeHash returns the code hash of the address
func (sd *ViewDB) GetCodeHash(addr common.Address) hash.Hash256 {
	return sd.GetState(addr, KeywordCodeHash)
}

// GetCode returns the code of the address
func (sd *ViewDB) GetCode(addr common.Address) []byte {
	return sd.Loader.AccountData(addr, KeywordCode[:])
}

// SetCode is not allowed
func (sd *ViewDB) SetCode(addr common.Address, code []byte) {
	panic(ErrNotAllowed)
}

// GetCodeSize returns the code size of the address
func (sd *ViewDB) GetCodeSize(addr common.Address) int {
	bs := sd.Loader.AccountData(addr, KeywordCodeSize[:])
	var Len int
	if len(bs) == 4 {
		Len = int(binary.LittleEndian.Uint32(bs))
	}
	return Len
}

// GetState returns value by the hash of the address
func (sd *ViewDB) GetState(addr common.Address, h hash.Hash256) hash.Hash256 {
	var ret hash.Hash256
	bs := sd.Loader.AccountData(addr, h[:])
	if len(bs) > 0 {
		copy(ret[:], bs)
	}
	return ret
}

// SetState is not allowed
func (sd *ViewDB) SetState(addr common.Address, h hash.Hash256, v hash.Hash256) {
	panic(ErrNotAllowed)
}

// Suicide is not allowed
func (sd *ViewDB) Suicide(addr common.Address) bool {
	panic(ErrNotAllowed)
	return false
}

// HasSuicided checks the dead state of the address
func (sd *ViewDB) HasSuicided(addr common.Address) bool {
	bs := sd.Loader.AccountData(addr, KeywordSuicide[:])
	return len(bs) > 0 && bs[0] == 1
}

// Exist checks that the account of the address is exist or not
func (sd *ViewDB) Exist(addr common.Address) bool {
	if exist, err := sd.Loader.IsExistAccount(addr); err != nil {
		panic(err)
	} else {
		return exist
	}
}

// Empty checks that seq == 0, balance == 0, code size == 0
func (sd *ViewDB) Empty(addr common.Address) bool {
	if acc, err := sd.Loader.Account(addr); err != nil {
		if err != data.ErrNotExistAccount {
			panic(err)
		}
		return true
	} else {
		balance := acc.Balance(sd.ChainCoord)
		return sd.Loader.Seq(addr) == 0 && balance.IsZero() && sd.GetCodeSize(addr) == 0
	}
}

// RevertToSnapshot doesn't work
func (sd *ViewDB) RevertToSnapshot(n int) {
}

// CommitSnapshot doesn't work
func (sd *ViewDB) CommitSnapshot(n int) {
}

// Snapshot doesn't work
func (sd *ViewDB) Snapshot() int {
	return 0
}

// AddLog not implemented yet
func (sd *ViewDB) AddLog(l *vm.Log) {
	log.Println("AddLog", l)
}

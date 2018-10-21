package solidity

import (
	"encoding/binary"
	"log"

	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/common/hash"
	"git.fleta.io/fleta/core/amount"
	"git.fleta.io/fleta/core/data"
	"git.fleta.io/fleta/core/db"
	"git.fleta.io/fleta/solidity/vm"
)

// ViewDB errors
var ()

// ViewDB is an EVM database for full state querying.
type ViewDB struct {
	ChainCoord *common.Coordinate
	Loader     data.Loader
}

// CreateAccount TODO
func (sd *ViewDB) CreateAccount(addr common.Address) {
	panic(ErrNotAllowed)
}

// SubBalance TODO
func (sd *ViewDB) SubBalance(addr common.Address, b *amount.Amount) {
	panic(ErrNotAllowed)
}

//AddBalance TODO
func (sd *ViewDB) AddBalance(addr common.Address, b *amount.Amount) {
	panic(ErrNotAllowed)
}

// GetBalance TODO
func (sd *ViewDB) GetBalance(addr common.Address) *amount.Amount {
	acc, err := sd.Loader.Account(addr)
	if err != nil {
		panic(err)
	}
	return acc.Balance(sd.ChainCoord)
}

// GetSeq TODO
func (sd *ViewDB) GetSeq(addr common.Address) uint64 {
	return sd.Loader.Seq(addr)
}

// AddSeq TODO
func (sd *ViewDB) AddSeq(addr common.Address) {
	panic(ErrNotAllowed)
}

// GetCodeHash TODO
func (sd *ViewDB) GetCodeHash(addr common.Address) hash.Hash256 {
	return sd.GetState(addr, KeywordCodeHash)
}

// GetCode TODO
func (sd *ViewDB) GetCode(addr common.Address) []byte {
	return sd.Loader.AccountData(addr, KeywordCode[:])
}

// SetCode TODO
func (sd *ViewDB) SetCode(addr common.Address, code []byte) {
	panic(ErrNotAllowed)
}

// GetCodeSize TODO
func (sd *ViewDB) GetCodeSize(addr common.Address) int {
	bs := sd.Loader.AccountData(addr, KeywordCodeSize[:])
	var Len int
	if len(bs) == 4 {
		Len = int(binary.LittleEndian.Uint32(bs))
	}
	return Len
}

// GetState TODO
func (sd *ViewDB) GetState(addr common.Address, h hash.Hash256) hash.Hash256 {
	var ret hash.Hash256
	bs := sd.Loader.AccountData(addr, h[:])
	if len(bs) > 0 {
		copy(ret[:], bs)
	}
	return ret
}

// SetState TODO
func (sd *ViewDB) SetState(addr common.Address, h hash.Hash256, v hash.Hash256) {
	panic(ErrNotAllowed)
}

// Suicide TODO
func (sd *ViewDB) Suicide(addr common.Address) bool {
	panic(ErrNotAllowed)
	return false
}

// HasSuicided TODO
func (sd *ViewDB) HasSuicided(addr common.Address) bool {
	bs := sd.Loader.AccountData(addr, KeywordSuicide[:])
	return len(bs) > 0 && bs[0] == 1
}

// Exist TODO
func (sd *ViewDB) Exist(addr common.Address) bool {
	if exist, err := sd.Loader.IsExistAccount(addr); err != nil {
		panic(err)
	} else {
		return exist
	}
}

// Empty TODO
func (sd *ViewDB) Empty(addr common.Address) bool {
	if acc, err := sd.Loader.Account(addr); err != nil {
		if err != db.ErrNotExistKey {
			panic(err)
		}
		return true
	} else {
		balance := acc.Balance(sd.ChainCoord)
		return sd.Loader.Seq(addr) == 0 && balance.IsZero() && sd.GetCodeSize(addr) == 0
	}
}

// RevertToSnapshot TODO
func (sd *ViewDB) RevertToSnapshot(n int) {
}

// CommitSnapshot TODO
func (sd *ViewDB) CommitSnapshot(n int) {
}

// Snapshot TODO
func (sd *ViewDB) Snapshot() int {
	return 0
}

// AddLog TODO
func (sd *ViewDB) AddLog(l *vm.Log) {
	log.Println("AddLog", l)
}

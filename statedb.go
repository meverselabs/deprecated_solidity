package solidity

import (
	"encoding/binary"
	"log"

	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/common/hash"
	"git.fleta.io/fleta/core/accounter"
	"git.fleta.io/fleta/core/amount"
	"git.fleta.io/fleta/core/data"
	"git.fleta.io/fleta/core/db"
	"git.fleta.io/fleta/solidity/vm"
)

// keywords StateDB
var (
	KeywordCode     = hash.Hash([]byte("__CODE__"))
	KeywordCodeHash = hash.Hash([]byte("__CODEHASH__"))
	KeywordCodeSize = hash.Hash([]byte("__CODESIZE__"))
	KeywordSuicide  = hash.Hash([]byte("__SUICIDE__"))
	KeywordMap      = map[hash.Hash256]bool{}
)

func init() {
	KeywordMap[KeywordCode] = true
	KeywordMap[KeywordCodeHash] = true
	KeywordMap[KeywordSuicide] = true
}

// StateDB is an EVM database for full state querying.
type StateDB struct {
	ChainCoord *common.Coordinate
	Context    *data.Context
}

// CreateAccount TODO
func (sd *StateDB) CreateAccount(addr common.Address) {
	//log.Println("CreateAccount", addr)
	act, err := accounter.ByCoord(sd.ChainCoord)
	if err != nil {
		panic(err)
	}
	a, err := act.NewByTypeName("solidity.Account")
	if err != nil {
		panic(err)
	}
	acc := a.(*Account)
	acc.Address_ = addr
	if err := sd.Context.CreateAccount(acc); err != nil {
		panic(err)
	}
}

// SubBalance TODO
func (sd *StateDB) SubBalance(addr common.Address, b *amount.Amount) {
	//log.Println("SubBalance", addr, b)
	acc, err := sd.Context.Account(addr)
	if err != nil {
		panic(err)
	}
	balance := acc.Balance(sd.ChainCoord)
	if balance.Less(b) {
		panic("")
	}
	balance = balance.Sub(b)
	acc.SetBalance(sd.ChainCoord, balance)
}

//AddBalance TODO
func (sd *StateDB) AddBalance(addr common.Address, b *amount.Amount) {
	//log.Println("AddBalance", addr, b)
	acc, err := sd.Context.Account(addr)
	if err != nil {
		panic(err)
	}
	balance := acc.Balance(sd.ChainCoord)
	balance = balance.Add(b)
	acc.SetBalance(sd.ChainCoord, balance)
}

// GetBalance TODO
func (sd *StateDB) GetBalance(addr common.Address) *amount.Amount {
	//log.Println("GetBalance", addr)
	acc, err := sd.Context.Account(addr)
	if err != nil {
		panic(err)
	}
	return acc.Balance(sd.ChainCoord)
}

// GetSeq TODO
func (sd *StateDB) GetSeq(addr common.Address) uint64 {
	//log.Println("GetSeq", addr)
	return sd.Context.Seq(addr)
}

// AddSeq TODO
func (sd *StateDB) AddSeq(addr common.Address) {
	//log.Println("AddSeq", addr)
	sd.Context.AddSeq(addr)
}

// GetCodeHash TODO
func (sd *StateDB) GetCodeHash(addr common.Address) hash.Hash256 {
	//log.Println("GetCodeHash", addr)
	return sd.GetState(addr, KeywordCodeHash)
}

// GetCode TODO
func (sd *StateDB) GetCode(addr common.Address) []byte {
	//log.Println("GetCode", addr)
	return sd.Context.AccountData(addr, KeywordCode[:])
}

// SetCode TODO
func (sd *StateDB) SetCode(addr common.Address, code []byte) {
	//log.Println("SetCode", addr, code)
	sd.Context.SetAccountData(addr, KeywordCode[:], code)
	h := hash.Hash(code)
	sd.Context.SetAccountData(addr, KeywordCodeHash[:], h[:])
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(len(code)))
	sd.Context.SetAccountData(addr, KeywordCodeSize[:], bs)
}

// GetCodeSize TODO
func (sd *StateDB) GetCodeSize(addr common.Address) int {
	//log.Println("GetCodeSize", addr)
	bs := sd.Context.AccountData(addr, KeywordCodeSize[:])
	var Len int
	if len(bs) == 4 {
		Len = int(binary.LittleEndian.Uint32(bs))
	}
	return Len
}

// GetState TODO
func (sd *StateDB) GetState(addr common.Address, h hash.Hash256) hash.Hash256 {
	//log.Println("GetState", addr, h)
	var ret hash.Hash256
	bs := sd.Context.AccountData(addr, h[:])
	if len(bs) > 0 {
		copy(ret[:], bs)
	}
	return ret
}

// SetState TODO
func (sd *StateDB) SetState(addr common.Address, h hash.Hash256, v hash.Hash256) {
	//log.Println("SetState", addr, h, v)
	if KeywordMap[h] {
		panic("reserved keyword")
	}
	sd.Context.SetAccountData(addr, h[:], v[:])
}

// Suicide TODO
func (sd *StateDB) Suicide(addr common.Address) bool {
	//log.Println("Suicide", addr)
	sd.Context.SetAccountData(addr, KeywordSuicide[:], []byte{1})
	return true
}

// HasSuicided TODO
func (sd *StateDB) HasSuicided(addr common.Address) bool {
	//log.Println("HasSuicided", addr)
	bs := sd.Context.AccountData(addr, KeywordSuicide[:])
	return len(bs) > 0 && bs[0] == 1
}

// Exist TODO
func (sd *StateDB) Exist(addr common.Address) bool {
	//log.Println("Exist", addr)
	if exist, err := sd.Context.IsExistAccount(addr); err != nil {
		panic(err)
	} else {
		return exist
	}
}

// Empty TODO
func (sd *StateDB) Empty(addr common.Address) bool {
	//log.Println("Empty", addr)
	if acc, err := sd.Context.Account(addr); err != nil {
		if err != db.ErrNotExistKey {
			panic(err)
		}
		return true
	} else {
		balance := acc.Balance(sd.ChainCoord)
		return sd.Context.Seq(addr) == 0 && balance.IsZero() && sd.GetCodeSize(addr) == 0
	}
}

// RevertToSnapshot TODO
func (sd *StateDB) RevertToSnapshot(n int) {
	//log.Println("RevertToSnapshot", n)
	sd.Context.Revert(n)
}

// CommitSnapshot TODO
func (sd *StateDB) CommitSnapshot(n int) {
	//log.Println("CommitSnapshot", n)
	sd.Context.Commit(n)
}

// Snapshot TODO
func (sd *StateDB) Snapshot() int {
	n := sd.Context.Snapshot()
	//log.Println("Snapshot", n)
	return n
}

// AddLog TODO
func (sd *StateDB) AddLog(l *vm.Log) {
	log.Println("AddLog", l)
}

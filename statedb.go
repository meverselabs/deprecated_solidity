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

// CreateAccount creates the sub account of the address to the context inside of EVM
func (sd *StateDB) CreateAccount(addr common.Address) {
	//log.Println("CreateAccount", addr)
	a, err := sd.Context.Accounter().NewByTypeName("solidity.Account")
	if err != nil {
		panic(err)
	}
	acc := a.(*Account)
	acc.Address_ = addr
	if err := sd.Context.CreateAccount(acc); err != nil {
		panic(err)
	}
}

// SubBalance reduce the target chain balance from the account of the address
func (sd *StateDB) SubBalance(addr common.Address, b *amount.Amount) {
	//log.Println("SubBalance", addr, b)
	balance, err := sd.Context.AccountBalance(addr)
	if err != nil {
		panic(err)
	}
	if err := balance.SubBalance(sd.ChainCoord, b); err != nil {
		panic(err)
	}
}

// AddBalance add the target chain balance to the account of the address
func (sd *StateDB) AddBalance(addr common.Address, b *amount.Amount) {
	//log.Println("AddBalance", addr, b)
	balance, err := sd.Context.AccountBalance(addr)
	if err != nil {
		panic(err)
	}
	balance.AddBalance(sd.ChainCoord, b)
}

// GetBalance returns the target chain balance from the account of the address
func (sd *StateDB) GetBalance(addr common.Address) *amount.Amount {
	//log.Println("GetBalance", addr)
	balance, err := sd.Context.AccountBalance(addr)
	if err != nil {
		panic(err)
	}
	return balance.Balance(sd.ChainCoord)
}

// GetSeq returns the sequence of the address
func (sd *StateDB) GetSeq(addr common.Address) uint64 {
	//log.Println("GetSeq", addr)
	return sd.Context.Seq(addr)
}

// AddSeq adds the sequence of the address
func (sd *StateDB) AddSeq(addr common.Address) {
	//log.Println("AddSeq", addr)
	sd.Context.AddSeq(addr)
}

// GetCodeHash returns the code hash of the address
func (sd *StateDB) GetCodeHash(addr common.Address) hash.Hash256 {
	//log.Println("GetCodeHash", addr)
	return sd.GetState(addr, KeywordCodeHash)
}

// GetCode returns the code of the address
func (sd *StateDB) GetCode(addr common.Address) []byte {
	//log.Println("GetCode", addr)
	return sd.Context.AccountData(addr, KeywordCode[:])
}

// SetCode updates the code to the address
func (sd *StateDB) SetCode(addr common.Address, code []byte) {
	//log.Println("SetCode", addr, code)
	sd.Context.SetAccountData(addr, KeywordCode[:], code)
	h := hash.Hash(code)
	sd.Context.SetAccountData(addr, KeywordCodeHash[:], h[:])
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(len(code)))
	sd.Context.SetAccountData(addr, KeywordCodeSize[:], bs)
}

// GetCodeSize returns the code size of the address
func (sd *StateDB) GetCodeSize(addr common.Address) int {
	//log.Println("GetCodeSize", addr)
	bs := sd.Context.AccountData(addr, KeywordCodeSize[:])
	var Len int
	if len(bs) == 4 {
		Len = int(binary.LittleEndian.Uint32(bs))
	}
	return Len
}

// GetState returns value by the hash of the address
func (sd *StateDB) GetState(addr common.Address, h hash.Hash256) hash.Hash256 {
	//log.Println("GetState", addr, h)
	var ret hash.Hash256
	bs := sd.Context.AccountData(addr, h[:])
	if len(bs) > 0 {
		copy(ret[:], bs)
	}
	return ret
}

// SetState updates value by the hash of the address
func (sd *StateDB) SetState(addr common.Address, h hash.Hash256, v hash.Hash256) {
	//log.Println("SetState", addr, h, v)
	if KeywordMap[h] {
		panic("reserved keyword")
	}
	sd.Context.SetAccountData(addr, h[:], v[:])
}

// Suicide make the address to dead state
func (sd *StateDB) Suicide(addr common.Address) bool {
	//log.Println("Suicide", addr)
	sd.Context.SetAccountData(addr, KeywordSuicide[:], []byte{1})
	return true
}

// HasSuicided checks the dead state of the address
func (sd *StateDB) HasSuicided(addr common.Address) bool {
	//log.Println("HasSuicided", addr)
	bs := sd.Context.AccountData(addr, KeywordSuicide[:])
	return len(bs) > 0 && bs[0] == 1
}

// Exist checks that the account of the address is exist or not
func (sd *StateDB) Exist(addr common.Address) bool {
	//log.Println("Exist", addr)
	if exist, err := sd.Context.IsExistAccount(addr); err != nil {
		panic(err)
	} else {
		return exist
	}
}

// Empty checks that seq == 0, balance == 0, code size == 0
func (sd *StateDB) Empty(addr common.Address) bool {
	//log.Println("Empty", addr)
	balance, err := sd.Context.AccountBalance(addr)
	if err != nil {
		panic(err)
	}
	return sd.Context.Seq(addr) == 0 && balance.Balance(sd.ChainCoord).IsZero() && sd.GetCodeSize(addr) == 0
}

// RevertToSnapshot removes snapshots after the snapshot number
func (sd *StateDB) RevertToSnapshot(n int) {
	//log.Println("RevertToSnapshot", n)
	sd.Context.Revert(n)
}

// CommitSnapshot apply snapshots to the top after the snapshot number
func (sd *StateDB) CommitSnapshot(n int) {
	//log.Println("CommitSnapshot", n)
	sd.Context.Commit(n)
}

// Snapshot push a snapshot and returns the snapshot number of it
func (sd *StateDB) Snapshot() int {
	n := sd.Context.Snapshot()
	//log.Println("Snapshot", n)
	return n
}

// AddLog not implemented yet
func (sd *StateDB) AddLog(l *vm.Log) {
	log.Println("AddLog", l)
}

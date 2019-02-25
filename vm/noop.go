// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package vm

import (
	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/common/hash"
	"git.fleta.io/fleta/core/amount"
)

func NoopCanTransfer(db StateDB, from common.Address, balance *amount.Amount) bool {
	return true
}
func NoopTransfer(db StateDB, from, to common.Address, amount *amount.Amount) {}

type NoopEVMCallContext struct{}

func (NoopEVMCallContext) Call(caller ContractRef, addr common.Address, data []byte, value *amount.Amount) ([]byte, error) {
	return nil, nil
}
func (NoopEVMCallContext) CallCode(caller ContractRef, addr common.Address, data []byte, value *amount.Amount) ([]byte, error) {
	return nil, nil
}
func (NoopEVMCallContext) Create(caller ContractRef, data []byte, value *amount.Amount) ([]byte, common.Address, error) {
	return nil, common.Address{}, nil
}
func (NoopEVMCallContext) DelegateCall(me ContractRef, addr common.Address, data []byte) ([]byte, error) {
	return nil, nil
}

type NoopStateDB struct{}

func (NoopStateDB) CreateAccount(common.Address, string)                {}
func (NoopStateDB) SubBalance(common.Address, *amount.Amount)           {}
func (NoopStateDB) AddBalance(common.Address, *amount.Amount)           {}
func (NoopStateDB) GetBalance(common.Address) *amount.Amount            { return nil }
func (NoopStateDB) GetNonce(common.Address) uint64                      { return 0 }
func (NoopStateDB) AddSeq(common.Address)                               {}
func (NoopStateDB) GetCodeHash(common.Address) hash.Hash256             { return hash.Hash256{} }
func (NoopStateDB) GetCode(common.Address) []byte                       { return nil }
func (NoopStateDB) SetCode(common.Address, []byte)                      {}
func (NoopStateDB) GetCodeSize(common.Address) int                      { return 0 }
func (NoopStateDB) GetState(common.Address, hash.Hash256) hash.Hash256  { return hash.Hash256{} }
func (NoopStateDB) SetState(common.Address, hash.Hash256, hash.Hash256) {}
func (NoopStateDB) Suicide(common.Address) bool                         { return false }
func (NoopStateDB) HasSuicided(common.Address) bool                     { return false }
func (NoopStateDB) Exist(common.Address) bool                           { return false }
func (NoopStateDB) Empty(common.Address) bool                           { return false }
func (NoopStateDB) RevertToSnapshot(int)                                {}
func (NoopStateDB) Snapshot() int                                       { return 0 }
func (NoopStateDB) AddLog(*Log)                                         {}

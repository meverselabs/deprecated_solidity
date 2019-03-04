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
	"github.com/fletaio/common"
	"github.com/fletaio/common/hash"
	"github.com/fletaio/core/amount"
)

// StateDB is an EVM database for full state querying.
type StateDB interface {
	CreateAccount(common.Address, string)

	SubBalance(common.Address, *amount.Amount)
	AddBalance(common.Address, *amount.Amount)
	GetBalance(common.Address) *amount.Amount

	GetSeq(common.Address) uint64
	AddSeq(common.Address)

	GetCodeHash(common.Address) hash.Hash256
	GetCode(common.Address) []byte
	SetCode(common.Address, []byte)
	GetCodeSize(common.Address) int

	GetState(common.Address, hash.Hash256) hash.Hash256
	SetState(common.Address, hash.Hash256, hash.Hash256)

	Suicide(common.Address) bool
	HasSuicided(common.Address) bool

	// Exist reports whether the given account exists in state.
	// Notably this should also return true for suicided accounts.
	Exist(common.Address) bool
	// Empty returns whether the given account is empty. Empty
	// is defined according to EIP161 (balance = nonce = code = 0).
	Empty(common.Address) bool

	RevertToSnapshot(int)
	CommitSnapshot(int)
	Snapshot() int

	AddLog(*Log)
}

// CallContext provides a basic interface for the EVM calling conventions. The EVM EVM
// depends on this context being implemented for doing subcalls and initialising new EVM contracts.
type CallContext interface {
	// Call another contract
	Call(env *EVM, me ContractRef, addr common.Address, data []byte, value *amount.Amount) ([]byte, error)
	// Take another's contract code and execute within our own context
	CallCode(env *EVM, me ContractRef, addr common.Address, data []byte, value *amount.Amount) ([]byte, error)
	// Same as CallCode except sender and value is propagated from parent to child scope
	DelegateCall(env *EVM, me ContractRef, addr common.Address, data []byte) ([]byte, error)
	// Create a new contract
	Create(env *EVM, me ContractRef, data []byte, value *amount.Amount) ([]byte, common.Address, error)
}

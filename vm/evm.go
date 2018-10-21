// Copyright 2014 The go-ethereum Authors
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
	"log"
	"math/big"
	"sync/atomic"
	"time"

	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/common/hash"
	"git.fleta.io/fleta/core/amount"
)

// emptyCodeHash is used by create to ensure deployment is disallowed to already
// deployed contract addresses (relevant after the account abstraction).
var emptyCodeHash = hash.Hash(nil)

type (
	CanTransferFunc func(StateDB, common.Address, *amount.Amount) bool
	TransferFunc    func(StateDB, common.Address, common.Address, *amount.Amount)
	// GetHashFunc returns the nth block hash in the blockchain
	// and is used by the BLOCKHASH EVM op code.
	GetHashFunc func(uint64) hash.Hash256
)

// run runs the given contract and takes care of running precompiles with a fallback to the byte code interpreter.
func run(evm *EVM, contract *Contract, input []byte) ([]byte, error) {
	if contract.CodeAddr != nil {
		precompiles := PrecompiledContractsByzantium
		if p := precompiles[*contract.CodeAddr]; p != nil {
			return RunPrecompiledContract(p, input, contract)
		}
	}
	return evm.interpreter.Run(contract, input)
}

// Context provides the EVM with auxiliary information. Once provided
// it shouldn't be modified.
type Context struct {
	// CanTransfer returns whether the account contains
	// sufficient ether to transfer the value
	CanTransfer CanTransferFunc
	// Transfer transfers ether from one account to the other
	Transfer TransferFunc
	// GetHash returns the hash corresponding to n
	GetHash GetHashFunc

	// Message information
	Origin common.Address // Provides information for ORIGIN

	// Block information
	Coinbase    common.Address // Provides information for COINBASE
	BlockNumber *big.Int       // Provides information for NUMBER
	Time        *big.Int       // Provides information for TIME
	Difficulty  *big.Int       // Provides information for DIFFICULTY
}

// EVM is the Ethereum Virtual Machine base object and provides
// the necessary tools to run a contract on the given state with
// the provided context. It should be noted that any error
// generated through any of the calls should be considered a
// revert-state-and-consume-all operation, no checks on
// specific errors should ever be performed. The interpreter makes
// sure that any errors generated are to be considered faulty code.
//
// The EVM should never be reused and is not thread safe.
type EVM struct {
	// Context provides auxiliary blockchain related information
	Context
	// StateDB gives access to the underlying state
	StateDB StateDB
	// Depth is the current call stack
	depth int

	// virtual machine configuration options used to initialise the
	// evm.
	vmConfig Config
	// global (to this context) ethereum virtual machine
	// used throughout the execution of the tx.
	interpreter *Interpreter
	// abort is used to abort the EVM calling operations
	// NOTE: must be set atomically
	abort int32
}

// NewEVM returns a new EVM. The returned EVM is not thread safe and should
// only ever be used *once*.
func NewEVM(ctx Context, statedb StateDB, vmConfig Config) *EVM {
	evm := &EVM{
		Context:  ctx,
		StateDB:  statedb,
		vmConfig: vmConfig,
	}

	evm.interpreter = NewInterpreter(evm, vmConfig)
	return evm
}

// Cancel cancels any running EVM operation. This may be called concurrently and
// it's safe to be called multiple times.
func (evm *EVM) Cancel() {
	atomic.StoreInt32(&evm.abort, 1)
}

// Call executes the contract associated with the addr with the given input as
// parameters. It also handles any necessary value transfer required and takes
// the necessary steps to create accounts and reverses the state in case of an
// execution error or failed value transfer.
func (evm *EVM) Call(caller ContractRef, addr common.Address, input []byte, value *amount.Amount) (ret []byte, err error) {
	if evm.vmConfig.NoRecursion && evm.depth > 0 {
		return nil, nil
	}

	// Fail if we're trying to execute above the call depth limit
	if evm.depth > int(CallCreateDepth) {
		return nil, ErrDepth
	}
	// Fail if we're trying to transfer more than the available balance
	if !evm.Context.CanTransfer(evm.StateDB, caller.Address(), value) {
		return nil, ErrInsufficientBalance
	}

	var (
		to       = AccountRef(addr)
		snapshot = evm.StateDB.Snapshot()
	)
	defer evm.StateDB.RevertToSnapshot(snapshot)
	if !evm.StateDB.Exist(addr) {
		return nil, ErrNotExistContract
	}
	evm.Transfer(evm.StateDB, caller.Address(), to.Address(), value)
	code := evm.StateDB.GetCode(addr)
	if len(code) == 0 {
		return nil, ErrInvalidContract
	}

	// Initialise a new contract and set the code that is to be used by the EVM.
	// The contract is a scoped environment for this execution context only.
	contract := NewContract(caller, to, value)
	contract.SetCallCode(&addr, evm.StateDB.GetCodeHash(addr), code)

	start := time.Now()

	// Capture the tracer start/end events in debug mode
	if evm.vmConfig.Debug && evm.depth == 0 {
		evm.vmConfig.Tracer.CaptureStart(caller.Address(), addr, false, input, value)

		defer func() { // Lazy evaluation of the parameters
			evm.vmConfig.Tracer.CaptureEnd(ret, time.Since(start), err)
		}()
	}
	ret, err = run(evm, contract, input)

	// When an error was returned by the EVM or when setting the creation code
	// above we revert to the snapshot and consume any remaining. Additionally
	// when we're in homestead this also counts for code storage errors.
	if err == nil {
		evm.StateDB.CommitSnapshot(snapshot)
	}
	return ret, err
}

// CallCode executes the contract associated with the addr with the given input
// as parameters. It also handles any necessary value transfer required and takes
// the necessary steps to create accounts and reverses the state in case of an
// execution error or failed value transfer.
//
// CallCode differs from Call in the sense that it executes the given address'
// code with the caller as context.
func (evm *EVM) CallCode(caller ContractRef, addr common.Address, input []byte, value *amount.Amount) (ret []byte, err error) {
	if evm.vmConfig.NoRecursion && evm.depth > 0 {
		return nil, nil
	}

	// Fail if we're trying to execute above the call depth limit
	if evm.depth > int(CallCreateDepth) {
		return nil, ErrDepth
	}
	// Fail if we're trying to transfer more than the available balance
	if !evm.CanTransfer(evm.StateDB, caller.Address(), value) {
		return nil, ErrInsufficientBalance
	}

	var (
		snapshot = evm.StateDB.Snapshot()
		to       = AccountRef(caller.Address())
	)
	defer evm.StateDB.RevertToSnapshot(snapshot)
	// initialise a new contract and set the code that is to be used by the
	// EVM. The contract is a scoped environment for this execution context
	// only.
	contract := NewContract(caller, to, value)
	contract.SetCallCode(&addr, evm.StateDB.GetCodeHash(addr), evm.StateDB.GetCode(addr))

	ret, err = run(evm, contract, input)
	if err == nil {
		evm.StateDB.CommitSnapshot(snapshot)
	}
	return ret, err
}

// DelegateCall executes the contract associated with the addr with the given input
// as parameters. It reverses the state in case of an execution error.
//
// DelegateCall differs from CallCode in the sense that it executes the given address'
// code with the caller as context and the caller is set to the caller of the caller.
func (evm *EVM) DelegateCall(caller ContractRef, addr common.Address, input []byte) (ret []byte, err error) {
	if evm.vmConfig.NoRecursion && evm.depth > 0 {
		return nil, nil
	}
	// Fail if we're trying to execute above the call depth limit
	if evm.depth > int(CallCreateDepth) {
		return nil, ErrDepth
	}

	var (
		snapshot = evm.StateDB.Snapshot()
		to       = AccountRef(caller.Address())
	)
	defer evm.StateDB.RevertToSnapshot(snapshot)

	// Initialise a new contract and make initialise the delegate values
	contract := NewContract(caller, to, nil).AsDelegate()
	contract.SetCallCode(&addr, evm.StateDB.GetCodeHash(addr), evm.StateDB.GetCode(addr))

	ret, err = run(evm, contract, input)
	if err == nil {
		evm.StateDB.CommitSnapshot(snapshot)
	}
	return ret, err
}

// StaticCall executes the contract associated with the addr with the given input
// as parameters while disallowing any modifications to the state during the call.
// Opcodes that attempt to perform such modifications will result in exceptions
// instead of performing the modifications.
func (evm *EVM) StaticCall(caller ContractRef, addr common.Address, input []byte) (ret []byte, err error) {
	if evm.vmConfig.NoRecursion && evm.depth > 0 {
		return nil, nil
	}
	// Fail if we're trying to execute above the call depth limit
	if evm.depth > int(CallCreateDepth) {
		return nil, ErrDepth
	}
	// Make sure the readonly is only set if we aren't in readonly yet
	// this makes also sure that the readonly flag isn't removed for
	// child calls.
	if !evm.interpreter.readOnly {
		evm.interpreter.readOnly = true
		defer func() { evm.interpreter.readOnly = false }()
	}

	var (
		to       = AccountRef(addr)
		snapshot = evm.StateDB.Snapshot()
	)
	defer evm.StateDB.RevertToSnapshot(snapshot)
	// Initialise a new contract and set the code that is to be used by the
	// EVM. The contract is a scoped environment for this execution context
	// only.
	contract := NewContract(caller, to, amount.NewCoinAmount(0, 0))
	contract.SetCallCode(&addr, evm.StateDB.GetCodeHash(addr), evm.StateDB.GetCode(addr))

	// When an error was returned by the EVM or when setting the creation code
	// above we revert to the snapshot and consume any remaining. Additionally
	// when we're in Homestead this also counts for code storage errors.
	ret, err = run(evm, contract, input)
	if err == nil {
		evm.StateDB.CommitSnapshot(snapshot)
	}
	return ret, err
}

// Create creates a new contract using code as deployment code.
func (evm *EVM) Create(caller ContractRef, contractAddr common.Address, code []byte, value *amount.Amount) (ret []byte, err error) {

	// Depth check execution. Fail if we're trying to execute above the
	// limit.
	if evm.depth > int(CallCreateDepth) {
		return nil, ErrDepth
	}
	if !evm.CanTransfer(evm.StateDB, caller.Address(), value) {
		return nil, ErrInsufficientBalance
	}
	if evm.StateDB.Exist(contractAddr) {
		return nil, ErrExistContract
	}

	contractHash := evm.StateDB.GetCodeHash(contractAddr)
	if evm.StateDB.GetSeq(contractAddr) != 0 || (contractHash != (hash.Hash256{}) && contractHash != emptyCodeHash) {
		log.Println(evm.StateDB.GetSeq(contractAddr), contractHash)
		return nil, ErrContractAddressCollision
	}
	// Create a new account on the state
	snapshot := evm.StateDB.Snapshot()
	defer evm.StateDB.RevertToSnapshot(snapshot)

	evm.StateDB.CreateAccount(contractAddr)
	evm.StateDB.AddSeq(contractAddr)
	evm.Transfer(evm.StateDB, caller.Address(), contractAddr, value)

	// initialise a new contract and set the code that is to be used by the
	// EVM. The contract is a scoped environment for this execution context
	// only.
	contract := NewContract(caller, AccountRef(contractAddr), value)
	contract.SetCallCode(&contractAddr, hash.Hash(code), code)

	if evm.vmConfig.NoRecursion && evm.depth > 0 {
		return nil, nil
	}

	if evm.vmConfig.Debug && evm.depth == 0 {
		evm.vmConfig.Tracer.CaptureStart(caller.Address(), contractAddr, true, code, value)
	}
	start := time.Now()

	ret, err = run(evm, contract, nil)

	// check whether the max code size has been exceeded
	maxCodeSizeExceeded := len(ret) > MaxCodeSize
	// if the contract creation ran successfully and no errors were returned
	if err == nil && !maxCodeSizeExceeded {
		evm.StateDB.SetCode(contractAddr, ret)
	}

	// When an error was returned by the EVM or when setting the creation code
	// above we revert to the snapshot and consume any remaining. Additionally
	// when we're in homestead this also counts for code storage errors.
	// Assign err if contract code size exceeds the max while the err is still empty.
	if maxCodeSizeExceeded && err == nil {
		err = errMaxCodeSizeExceeded
	}
	if evm.vmConfig.Debug && evm.depth == 0 {
		evm.vmConfig.Tracer.CaptureEnd(ret, time.Since(start), err)
	}
	if err == nil {
		evm.StateDB.CommitSnapshot(snapshot)
	}
	return ret, err
}

// Interpreter returns the EVM interpreter
func (evm *EVM) Interpreter() *Interpreter { return evm.interpreter }

package solidity

import (
	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/core/amount"
	"git.fleta.io/fleta/solidity/vm"
)

// CanTransfer TODO
func CanTransfer(db vm.StateDB, addr common.Address, amount *amount.Amount) bool {
	return !db.GetBalance(addr).Less(amount)
}

// Transfer subtracts amount from sender and adds amount to recipient using the given Db
func Transfer(db vm.StateDB, sender, recipient common.Address, amount *amount.Amount) {
	if !amount.IsZero() {
		db.SubBalance(sender, amount)
		db.AddBalance(recipient, amount)
	}
}

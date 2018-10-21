package solidity

import (
	"io"

	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/core/account"
	"git.fleta.io/fleta/core/accounter"
	"git.fleta.io/fleta/core/amount"
	"git.fleta.io/fleta/core/data"
)

func init() {
	accounter.RegisterHandler("solidity.Account", func(t account.Type) account.Account {
		return &Account{
			Base: account.Base{
				Type_:       t,
				BalanceHash: map[uint64]*amount.Amount{},
			},
		}
	}, func(loader data.Loader, a account.Account, signers []common.PublicHash) error {
		return ErrNotAllowed
	})
}

// Account TODO
type Account struct {
	account.Base
}

// Clone TODO
func (acc *Account) Clone() account.Account {
	balanceHash := map[uint64]*amount.Amount{}
	for k, v := range acc.BalanceHash {
		balanceHash[k] = v.Clone()
	}
	return &Account{
		Base: account.Base{
			Address_:    acc.Address_,
			Type_:       acc.Type_,
			BalanceHash: balanceHash,
		},
	}
}

// WriteTo TODO
func (acc *Account) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := acc.Base.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	return wrote, nil
}

// ReadFrom TODO
func (acc *Account) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if n, err := acc.Base.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	return read, nil
}

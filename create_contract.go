package solidity

import (
	"bytes"
	"io"
	"math/big"
	"time"

	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/common/hash"
	"git.fleta.io/fleta/common/util"
	"git.fleta.io/fleta/core/accounter"
	"git.fleta.io/fleta/core/amount"
	"git.fleta.io/fleta/core/data"
	"git.fleta.io/fleta/core/transaction"
	"git.fleta.io/fleta/core/transactor"
	"git.fleta.io/fleta/solidity/vm"
)

func init() {
	transactor.RegisterHandler("solidity.CreateContract", func(t transaction.Type) transaction.Transaction {
		return &CreateContract{
			Base: transaction.Base{
				ChainCoord_: &common.Coordinate{},
				Type_:       t,
			},
		}
	}, func(loader data.Loader, t transaction.Transaction, signers []common.PublicHash) error {
		tx := t.(*CreateContract)
		if tx.Seq() <= loader.Seq(tx.From()) {
			return ErrInvalidSequence
		}

		fromAcc, err := loader.Account(tx.From())
		if err != nil {
			return err
		}

		act, err := accounter.ByCoord(loader.ChainCoord())
		if err != nil {
			return err
		}
		if err := act.Validate(fromAcc, signers); err != nil {
			return err
		}
		return nil
	}, func(Context *data.Context, Fee *amount.Amount, t transaction.Transaction, coord *common.Coordinate) (ret interface{}, rerr error) {
		defer func() {
			if e := recover(); e != nil {
				if err, is := e.(error); is {
					rerr = err
				} else {
					rerr = ErrVirtualMachinePanic
				}
			}
		}()

		tx := t.(*CreateContract)
		sn := Context.Snapshot()
		defer Context.Revert(sn)

		if tx.Seq() != Context.Seq(tx.From())+1 {
			return nil, ErrInvalidSequence
		}
		Context.AddSeq(tx.From())

		fromAcc, err := Context.Account(tx.From())
		if err != nil {
			return nil, err
		}

		chainCoord := Context.ChainCoord()
		balance := fromAcc.Balance(chainCoord)
		if balance.Less(Fee) {
			return nil, ErrInsuffcientBalance
		}
		balance = balance.Sub(Fee)
		fromAcc.SetBalance(chainCoord, balance)

		contAddr := common.NewAddress(coord, chainCoord, 0)
		if is, err := Context.IsExistAccount(contAddr); err != nil {
			return nil, err
		} else if is {
			return nil, ErrExistAddress
		}
		statedb := &StateDB{
			ChainCoord: chainCoord,
			Context:    Context,
		}
		logconfig := &vm.LogConfig{
			DisableMemory: false,
			DisableStack:  false,
			Debug:         false,
		}
		vmCfg := vm.Config{
			Tracer: vm.NewStructLogger(logconfig),
			Debug:  false,
		}
		vContext := vm.Context{
			CanTransfer: CanTransfer,
			Transfer:    Transfer,
			GetHash:     func(uint64) hash.Hash256 { return hash.Hash256{} },
			Origin:      tx.From(),
			BlockNumber: new(big.Int).SetUint64(100),
			Time:        big.NewInt(time.Now().Unix()),
			Difficulty:  new(big.Int),
		}
		evm := vm.NewEVM(vContext, statedb, vmCfg)
		code, err := evm.Create(vm.AccountRef(tx.From()), contAddr, append(tx.Code, tx.Params...), amount.NewCoinAmount(0, 0))
		if err != nil {
			return nil, err
		}
		Context.Commit(sn)
		return code, nil
	})
}

// CreateContract TODO
type CreateContract struct {
	transaction.Base
	Seq_   uint64
	From_  common.Address
	Code   []byte
	Params []byte
}

// IsUTXO TODO
func (tx *CreateContract) IsUTXO() bool {
	return false
}

// From TODO
func (tx *CreateContract) From() common.Address {
	return tx.From_
}

// Seq TODO
func (tx *CreateContract) Seq() uint64 {
	return tx.Seq_
}

// Hash TODO
func (tx *CreateContract) Hash() hash.Hash256 {
	var buffer bytes.Buffer
	if _, err := tx.WriteTo(&buffer); err != nil {
		panic(err)
	}
	return hash.DoubleHash(buffer.Bytes())
}

// WriteTo TODO
func (tx *CreateContract) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := tx.Base.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint64(w, tx.Seq_); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := tx.From_.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteBytes(w, tx.Code); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteBytes(w, tx.Params); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	return wrote, nil
}

// ReadFrom TODO
func (tx *CreateContract) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if n, err := tx.Base.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if v, n, err := util.ReadUint64(r); err != nil {
		return read, err
	} else {
		read += n
		tx.Seq_ = v
	}
	if n, err := tx.From_.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if bs, n, err := util.ReadBytes(r); err != nil {
		return read, err
	} else {
		read += n
		tx.Code = bs
	}
	if bs, n, err := util.ReadBytes(r); err != nil {
		return read, err
	} else {
		read += n
		tx.Params = bs
	}
	return read, nil
}

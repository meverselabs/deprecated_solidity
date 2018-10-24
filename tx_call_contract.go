package solidity

import (
	"bytes"
	"io"
	"math/big"
	"time"

	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/common/hash"
	"git.fleta.io/fleta/common/util"
	"git.fleta.io/fleta/core/amount"
	"git.fleta.io/fleta/core/data"
	"git.fleta.io/fleta/core/transaction"
	"git.fleta.io/fleta/core/transactor"
	"git.fleta.io/fleta/solidity/vm"
)

func init() {
	transactor.RegisterHandler("solidity.CallContract", func(t transaction.Type) transaction.Transaction {
		return &CallContract{
			Base: transaction.Base{
				ChainCoord_: &common.Coordinate{},
				Type_:       t,
			},
		}
	}, func(loader data.Loader, t transaction.Transaction, signers []common.PublicHash) error {
		tx := t.(*CallContract)
		if tx.Seq() <= loader.Seq(tx.From()) {
			return ErrInvalidSequence
		}

		fromAcc, err := loader.Account(tx.From())
		if err != nil {
			return err
		}

		if err := loader.Accounter().Validate(loader, fromAcc, signers); err != nil {
			return err
		}
		return nil
	}, func(ctx *data.Context, Fee *amount.Amount, t transaction.Transaction, coord *common.Coordinate) (ret interface{}, rerr error) {
		defer func() {
			if e := recover(); e != nil {
				if err, is := e.(error); is {
					rerr = err
				} else {
					rerr = ErrVirtualMachinePanic
				}
			}
		}()

		tx := t.(*CallContract)
		sn := ctx.Snapshot()
		defer ctx.Revert(sn)

		if tx.Seq() != ctx.Seq(tx.From())+1 {
			return nil, ErrInvalidSequence
		}
		ctx.AddSeq(tx.From())

		fromAcc, err := ctx.Account(tx.From())
		if err != nil {
			return nil, err
		}

		chainCoord := ctx.ChainCoord()
		balance := fromAcc.Balance(chainCoord)
		if balance.Less(Fee) {
			return nil, ErrInsuffcientBalance
		}
		balance = balance.Sub(Fee)
		fromAcc.SetBalance(chainCoord, balance)

		statedb := &StateDB{
			ChainCoord: chainCoord,
			Context:    ctx,
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
		vctx := vm.Context{
			CanTransfer: CanTransfer,
			Transfer:    Transfer,
			GetHash:     func(uint64) hash.Hash256 { return hash.Hash256{} },
			Origin:      tx.From(),
			BlockNumber: new(big.Int).SetUint64(100),
			Time:        big.NewInt(time.Now().Unix()),
			Difficulty:  new(big.Int),
		}
		evm := vm.NewEVM(vctx, statedb, vmCfg)
		ret, err = evm.Call(vm.AccountRef(tx.From()), tx.To, append(tx.Method, tx.Params...), amount.NewCoinAmount(0, 0))
		if err != nil {
			return nil, err
		}
		ctx.Commit(sn)
		return ret, nil
	})
}

// CallContract TODO
type CallContract struct {
	transaction.Base
	Seq_   uint64
	From_  common.Address
	To     common.Address
	Amount *amount.Amount
	Method []byte
	Params []byte
}

// IsUTXO TODO
func (tx *CallContract) IsUTXO() bool {
	return false
}

// From TODO
func (tx *CallContract) From() common.Address {
	return tx.From_
}

// Seq TODO
func (tx *CallContract) Seq() uint64 {
	return tx.Seq_
}

// Hash TODO
func (tx *CallContract) Hash() hash.Hash256 {
	var buffer bytes.Buffer
	if _, err := tx.WriteTo(&buffer); err != nil {
		panic(err)
	}
	return hash.DoubleHash(buffer.Bytes())
}

// WriteTo TODO
func (tx *CallContract) WriteTo(w io.Writer) (int64, error) {
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
	if n, err := tx.To.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := tx.Amount.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteBytes8(w, tx.Method); err != nil {
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
func (tx *CallContract) ReadFrom(r io.Reader) (int64, error) {
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
	if n, err := tx.To.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if n, err := tx.Amount.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if bs, n, err := util.ReadBytes8(r); err != nil {
		return read, err
	} else {
		read += n
		tx.Method = bs
	}
	if bs, n, err := util.ReadBytes(r); err != nil {
		return read, err
	} else {
		read += n
		tx.Params = bs
	}
	return read, nil
}

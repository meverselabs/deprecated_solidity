package solidity

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"math/big"
	"time"

	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/common/hash"
	"git.fleta.io/fleta/common/util"
	"git.fleta.io/fleta/core/amount"
	"git.fleta.io/fleta/core/data"
	"git.fleta.io/fleta/core/transaction"
	"git.fleta.io/fleta/solidity/vm"
)

func init() {
	data.RegisterTransaction("solidity.CallContract", func(t transaction.Type) transaction.Transaction {
		return &CallContract{
			Base: transaction.Base{
				Type_: t,
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
		if err := fromAcc.SubBalance(Fee); err != nil {
			return nil, err
		}

		statedb := &StateDB{
			Context: ctx,
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
		ret, err = evm.Call(vm.AccountRef(tx.From()), tx.To, append(tx.Method, tx.Params...), tx.Amount)
		if err != nil {
			return nil, err
		}
		ctx.Commit(sn)
		return ret, nil
	})
}

// CallContract is a solidity.CallContract
// It is used to call the contract method
type CallContract struct {
	transaction.Base
	Seq_   uint64
	From_  common.Address
	Amount *amount.Amount
	To     common.Address
	Method []byte
	Params []byte
}

// IsUTXO returns false
func (tx *CallContract) IsUTXO() bool {
	return false
}

// From returns the creator of the transaction
func (tx *CallContract) From() common.Address {
	return tx.From_
}

// Seq returns the sequence of the transaction
func (tx *CallContract) Seq() uint64 {
	return tx.Seq_
}

// Hash returns the hash value of it
func (tx *CallContract) Hash() hash.Hash256 {
	return hash.DoubleHashByWriterTo(tx)
}

// WriteTo is a serialization function
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
	if n, err := tx.Amount.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := tx.To.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteBytes(w, tx.Method); err != nil {
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

// ReadFrom is a deserialization function
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
	if n, err := tx.Amount.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if n, err := tx.To.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if bs, n, err := util.ReadBytes(r); err != nil {
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

// MarshalJSON is a marshaler function
func (tx *CallContract) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"type":`)
	if bs, err := json.Marshal(tx.Type_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"timestamp":`)
	if bs, err := json.Marshal(tx.Timestamp_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"seq":`)
	if bs, err := json.Marshal(tx.Seq_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"from":`)
	if bs, err := tx.From_.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"amount":`)
	if bs, err := tx.Amount.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"to":`)
	if bs, err := tx.To.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"method":`)
	if len(tx.Method) == 0 {
		buffer.WriteString(`null`)
	} else {
		buffer.WriteString(`"`)
		buffer.WriteString(hex.EncodeToString(tx.Method))
		buffer.WriteString(`"`)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"params":`)
	if len(tx.Params) == 0 {
		buffer.WriteString(`null`)
	} else {
		buffer.WriteString(`"`)
		buffer.WriteString(hex.EncodeToString(tx.Params))
		buffer.WriteString(`"`)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}

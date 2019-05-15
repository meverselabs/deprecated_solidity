package solidity

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"

	"github.com/fletaio/common/hash"
	"github.com/fletaio/common/util"
	"github.com/fletaio/core/data"
	"github.com/fletaio/core/event"

	"github.com/fletaio/common"
)

func init() {
	data.RegisterEvent("solidity.Log", func(t event.Type) event.Event {
		return &LogEvent{
			Base: event.Base{
				Type_: t,
			},
		}
	})
}

// LogEvent is a event of adding count to the account
type LogEvent struct {
	event.Base
	Address common.Address
	Topics  []hash.Hash256
	Data    []byte
	Removed bool `json:"removed"`
}

// WriteTo is a serialization function
func (e *LogEvent) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := e.Base.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := e.Address.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint8(w, uint8(len(e.Topics))); err != nil {
		return wrote, err
	} else {
		wrote += n
		for _, v := range e.Topics {
			if n, err := v.WriteTo(w); err != nil {
				return wrote, err
			} else {
				wrote += n
			}
		}
	}
	if n, err := util.WriteBytes(w, e.Data); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteBool(w, e.Removed); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	return wrote, nil
}

// ReadFrom is a deserialization function
func (e *LogEvent) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if n, err := e.Base.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if n, err := e.Address.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if Len, n, err := util.ReadUint8(r); err != nil {
		return read, err
	} else {
		read += n
		e.Topics = make([]hash.Hash256, Len)
		for i := 0; i < int(Len); i++ {
			var h hash.Hash256
			if n, err := h.ReadFrom(r); err != nil {
				return read, err
			} else {
				read += n
			}
			e.Topics = append(e.Topics, h)
		}
	}
	if bs, n, err := util.ReadBytes(r); err != nil {
		return read, err
	} else {
		read += n
		e.Data = bs
	}
	if v, n, err := util.ReadBool(r); err != nil {
		return read, err
	} else {
		read += n
		e.Removed = v
	}
	return read, nil
}

// MarshalJSON is a marshaler function
func (e *LogEvent) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"coord":`)
	if bs, err := e.Coord_.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"index":`)
	if bs, err := json.Marshal(e.Index_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"type":`)
	if bs, err := json.Marshal(e.Type_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"address":`)
	if bs, err := json.Marshal(e.Address); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"topics":`)
	buffer.WriteString(`[`)
	for i, h := range e.Topics {
		if i > 0 {
			buffer.WriteString(`,`)
		}
		if bs, err := h.MarshalJSON(); err != nil {
			return nil, err
		} else {
			buffer.Write(bs)
		}
	}
	buffer.WriteString(`]`)
	buffer.WriteString(`"data":`)
	if bs, err := json.Marshal(hex.EncodeToString(e.Data)); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"removed":`)
	if bs, err := json.Marshal(e.Removed); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}

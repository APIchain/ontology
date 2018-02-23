package states

import (
	"github.com/Ontology/common"
	"io"
	"github.com/Ontology/common/serialization"
	. "github.com/Ontology/errors"
)

type EventTxState struct {
	StateBase
	Txids []common.Uint256
}

func (this *EventTxState) Serialize(w io.Writer) error {
	this.StateBase.Serialize(w)
	err := serialization.WriteUint32(w, uint32(len(this.Txids)))
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "EventTxState Txids Len Serialize failed.")
	}
	for _, v := range this.Txids {
		v.Serialize(w)
	}
	return nil
}

func (this *EventTxState) Deserialize(r io.Reader) error {
	if this == nil {
		this = new(EventTxState)
	}

	err := this.StateBase.Deserialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "EventTxState StateBase Deserialize failed.")
	}

	n, err := serialization.ReadUint32(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "EventTxState Txids Len Deserialize failed.")
	}

	var u common.Uint256
	for i := 0; i < int(n); i++ {
		err = u.Deserialize(r)
		if err != nil {
			return err
		}
		this.Txids = append(this.Txids, u)
	}
	return nil
}
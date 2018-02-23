package ChainStore

import (
	. "github.com/Ontology/core/store"
	"github.com/Ontology/common"
	"github.com/Ontology/smartcontract/event"
	"encoding/json"
	"github.com/Ontology/core/states"
	"bytes"
	"github.com/Ontology/common/serialization"
)

const (
	EventDBDir = "Event"
)

var DefaultEventStore IEventStore

type IEventStore interface {
	SaveEventNotifyInTx(txid common.Uint256, notifies []*event.NotifyEventInfo) error
	SaveEventNotifyInBlock(height uint32, txids *states.EventTxState) error
	GetEventNotifyByTx(txid common.Uint256) ([]*event.NotifyEventInfo, error)
	GetEventNotifyTxIds(height uint32) (*states.EventTxState, error)
	BatchCommit() error
}

type EventStore struct {
	st IStore
}

func NewEventStore() (IEventStore, error) {
	st, err := NewStore(EventDBDir)
	if err != nil {
		return nil, err
	}
	return &EventStore{st}, nil
}

func (this *EventStore) SaveEventNotifyInTx(txid common.Uint256, notifies []*event.NotifyEventInfo) error {
	result, err := json.Marshal(notifies)
	if err != nil {
		return err
	}
	return this.st.BatchPut(append([]byte{byte(EVENT_Notify)}, txid.ToArray()...), result)
}

func (this *EventStore) GetEventNotifyByTx(txid common.Uint256) ([]*event.NotifyEventInfo, error) {
	result, err := this.st.Get(append([]byte{byte(EVENT_Notify)}, txid.ToArray()...))
	if err != nil {
		return nil, err
	}
	var notifies []*event.NotifyEventInfo
	if err = json.Unmarshal(result, &notifies); err != nil {
		return nil, err
	}
	return notifies, nil
}

func (this *EventStore) SaveEventNotifyInBlock(height uint32, txids *states.EventTxState) error {
	b := new(bytes.Buffer)
	if err := txids.Serialize(b); err != nil {
		return err
	}

	f := new(bytes.Buffer)
	if err := serialization.WriteUint32(f, height); err != nil {
		return err
	}
	return this.st.BatchPut(append([]byte{byte(EVENT_Notify)}, f.Bytes()...), b.Bytes())
}

func (this *EventStore) GetEventNotifyTxIds(height uint32) (*states.EventTxState, error) {
	f := new(bytes.Buffer)
	if err := serialization.WriteUint32(f, height); err != nil {
		return nil, err
	}
	result, err := this.st.Get(append([]byte{byte(EVENT_Notify)}, f.Bytes()...))
	if err != nil {
		return nil, err
	}
	txids := new(states.EventTxState)
	if err := txids.Deserialize(bytes.NewBuffer(result)); err != nil {
		return nil, err
	}
	return txids, nil
}

func (this *EventStore) BatchCommit() error {
	return this.st.BatchCommit()
}


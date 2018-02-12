package dbft

import (
	"math/big"
	"math/rand"
	"reflect"
	"sync"
	"testing"
	"time"

	cl "github.com/Ontology/account"
	. "github.com/Ontology/common"
	"github.com/Ontology/common/log"
	ser "github.com/Ontology/common/serialization"
	"github.com/Ontology/core/ledger"
	tx "github.com/Ontology/core/transaction"
	"github.com/Ontology/crypto"
	"github.com/Ontology/net"
	msg "github.com/Ontology/net/message"
)

func init() {
	log.Init(log.Path, log.Stdout)
	crypto.SetAlg("P256R1")
}

func GetBookeeps(num int) []*crypto.PubKey {
	var bookKeepers []*crypto.PubKey
	for i := 0; i < num; i++ {
		mPubKey := new(crypto.PubKey)
		mPubKey.X = big.NewInt(0)
		mPubKey.Y = big.NewInt(1)
		bookKeepers = append(bookKeepers, mPubKey)
	}
	return bookKeepers
}

func GetPubKey() *crypto.PubKey {
	mPubKey := new(crypto.PubKey)
	mPubKey.X = big.NewInt(0)
	mPubKey.Y = big.NewInt(1)
	return mPubKey
}

func GetPayload() *msg.ConsensusPayload {
	cv := &ChangeView{
		NewViewNumber: 'b',
	}
	cv.ConsensusMessageData().ViewNumber = 1
	return &msg.ConsensusPayload{
		Version:         ContextVersion,
		PrevHash:        *new(Uint256),
		Height:          3,
		BookKeeperIndex: uint16(4),
		Timestamp:       uint32(time.Now().UTC().UnixNano()),
		Data:            ser.ToArray(cv),
		Owner:           GetPubKey(),
	}
}
func GetMessage() ConsensusMessage {
	msg := &ChangeView{
		NewViewNumber: 1,
	}
	msg.msgData.Type = ChangeViewMsg
	msg.ConsensusMessageData().ViewNumber = 1
	return msg
}

func GetSignature(m int) [][]byte {
	signatures := make([][]byte, m)
	for i := 0; i < m; i++ {
		signatures[i] = []byte{'1', '2'}
	}
	return signatures
}

func TestConsensusContext_M(t *testing.T) {
	type fields struct {
		BookKeepers []*crypto.PubKey
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name:   "test",
			fields: fields{BookKeepers: GetBookeeps(4)},
			want:   3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cxt := &ConsensusContext{
				BookKeepers: tt.fields.BookKeepers,
			}
			if got := cxt.M(); got != tt.want {
				t.Errorf("ConsensusContext.M() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConsensusContext_ChangeView(t *testing.T) {
	type fields struct {
		State       ConsensusState
		Height      uint32
		BookKeepers []*crypto.PubKey
	}
	type args struct {
		viewNum byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "test",
			fields: fields{State: Primary, Height: 3, BookKeepers: GetBookeeps(4)},
			args:   args{viewNum: 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cxt := &ConsensusContext{
				State:       tt.fields.State,
				Height:      tt.fields.Height,
				BookKeepers: tt.fields.BookKeepers,
			}
			cxt.ChangeView(tt.args.viewNum)
		})
	}
}

func TestConsensusContext_MakeChangeView(t *testing.T) {
	type fields struct {
		PrevHash        Uint256
		Height          uint32
		ViewNumber      byte
		BookKeepers     []*crypto.PubKey
		Owner           *crypto.PubKey
		BookKeeperIndex int
		Timestamp       uint32
		ExpectedView    []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   *msg.ConsensusPayload
	}{
		{
			name: "test",
			fields: fields{PrevHash: *new(Uint256), Height: 1, ViewNumber: 1, BookKeepers: GetBookeeps(4), Owner: GetPubKey(),
				BookKeeperIndex: 1, Timestamp: uint32(time.Now().UTC().UnixNano()), ExpectedView: []byte("abc")},
			want: GetPayload(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cxt := &ConsensusContext{
				PrevHash:        tt.fields.PrevHash,
				Height:          tt.fields.Height,
				ViewNumber:      tt.fields.ViewNumber,
				BookKeepers:     tt.fields.BookKeepers,
				Owner:           tt.fields.Owner,
				BookKeeperIndex: tt.fields.BookKeeperIndex,
				Timestamp:       tt.fields.Timestamp,
				ExpectedView:    tt.fields.ExpectedView,
			}
			if got := cxt.MakeChangeView(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConsensusContext.MakeChangeView() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConsensusContext_MakeHeader(t *testing.T) {
	type fields struct {
		PrevHash        Uint256
		Height          uint32
		NextBookKeepers []*crypto.PubKey
		Timestamp       uint32
		Nonce           uint64
		Transactions    []*tx.Transaction
		header          *ledger.Block
	}
	tests := []struct {
		name   string
		fields fields
		want   *ledger.Block
	}{
		{
			name: "test",
			fields: fields{PrevHash: *new(Uint256), Height: 1, NextBookKeepers: GetBookeeps(2), Timestamp: uint32(time.Now().UTC().UnixNano()),
				Nonce: uint64(rand.Uint32())<<32 + uint64(rand.Uint32()), Transactions: nil, header: nil},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cxt := &ConsensusContext{
				PrevHash:        tt.fields.PrevHash,
				Height:          tt.fields.Height,
				NextBookKeepers: tt.fields.NextBookKeepers,
				Timestamp:       tt.fields.Timestamp,
				Nonce:           tt.fields.Nonce,
				Transactions:    tt.fields.Transactions,
				header:          tt.fields.header,
			}
			if got := cxt.MakeHeader(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConsensusContext.MakeHeader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConsensusContext_MakePayload(t *testing.T) {
	type fields struct {
		PrevHash        Uint256
		Height          uint32
		ViewNumber      byte
		Owner           *crypto.PubKey
		BookKeeperIndex int
		Timestamp       uint32
	}
	type args struct {
		message ConsensusMessage
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *msg.ConsensusPayload
	}{
		{
			name: "test",
			fields: fields{PrevHash: *new(Uint256), Height: 1, ViewNumber: 1, Owner: GetPubKey(),
				BookKeeperIndex: 1, Timestamp: uint32(time.Now().UTC().UnixNano())},
			args: args{message: GetMessage()},
			want: GetPayload(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cxt := &ConsensusContext{
				PrevHash:        tt.fields.PrevHash,
				Height:          tt.fields.Height,
				ViewNumber:      tt.fields.ViewNumber,
				Owner:           tt.fields.Owner,
				BookKeeperIndex: tt.fields.BookKeeperIndex,
				Timestamp:       tt.fields.Timestamp,
			}
			if got := cxt.MakePayload(tt.args.message); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConsensusContext.MakePayload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConsensusContext_GetSignaturesCount(t *testing.T) {
	type fields struct {
		Signatures [][]byte
	}
	tests := []struct {
		name      string
		fields    fields
		wantCount int
	}{
		{
			name:      "test",
			fields:    fields{Signatures: GetSignature(3)},
			wantCount: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cxt := &ConsensusContext{
				Signatures: tt.fields.Signatures,
			}
			if gotCount := cxt.GetSignaturesCount(); gotCount != tt.wantCount {
				t.Errorf("ConsensusContext.GetSignaturesCount() = %v, want %v", gotCount, tt.wantCount)
			}
		})
	}
}

func TestConsensusContext_GetStateDetail(t *testing.T) {
	instance := NewConsensusContext()
	instance.State = Primary
	res := instance.GetStateDetail()
	t.Log("TestGetStateDetail: ", res)
}


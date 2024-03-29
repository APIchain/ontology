package transaction

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	. "github.com/Ontology/common"
	. "github.com/Ontology/common/config"
	"github.com/Ontology/common/serialization"
	"github.com/Ontology/core/asset"
	"github.com/Ontology/core/contract"
	"github.com/Ontology/core/contract/program"
	sig "github.com/Ontology/core/signature"
	"github.com/Ontology/core/transaction/payload"
	. "github.com/Ontology/core/transaction/utxo"
	"github.com/Ontology/crypto"
	. "github.com/Ontology/errors"
	vm "github.com/Ontology/vm/neovm"
	"io"
	"math/big"
	"sort"
)

const (
	OntRegisterAmount = 1000000000
	OngRegisterAmount = 1000000000
)

var (
	Infinity    = &crypto.PubKey{X: big.NewInt(0), Y: big.NewInt(0)}
	SystemIssue Uint256
	ONTToken    = NewGoverningToken()
	ONGToken    = NewUtilityToken()
	ONTTokenID  = ONTToken.Hash()
	ONGTokenID  = ONGToken.Hash()
)

//for different transaction types with different payload format
//and transaction process methods
type TransactionType byte

const (
	BookKeeping    TransactionType = 0x00
	IssueAsset     TransactionType = 0x01
	BookKeeper     TransactionType = 0x02
	Claim          TransactionType = 0x03
	PrivacyPayload TransactionType = 0x20
	RegisterAsset  TransactionType = 0x40
	TransferAsset  TransactionType = 0x80
	Record         TransactionType = 0x81
	Deploy         TransactionType = 0xd0
	Invoke         TransactionType = 0xd1
	DataFile       TransactionType = 0x12
	Enrollment     TransactionType = 0x04
	Vote           TransactionType = 0x05
)

var TxName = map[TransactionType]string{
	BookKeeping:    "BookKeeping",
	IssueAsset:     "IssueAsset",
	BookKeeper:     "BookKeeper",
	Claim:          "Claim",
	PrivacyPayload: "PrivacyPayload",
	RegisterAsset:  "RegisterAsset",
	TransferAsset:  "TransferAsset",
	Record:         "Record",
	Deploy:         "Deploy",
	Invoke:         "Invoke",
	DataFile:       "DataFile",
	Enrollment:     "Enrollment",
	Vote:           "Vote",
}

//Payload define the func for loading the payload data
//base on payload type which have different struture
type Payload interface {
	//  Get payload data
	Data(version byte) []byte

	//Serialize payload data
	Serialize(w io.Writer, version byte) error

	Deserialize(r io.Reader, version byte) error
}

//Transaction is used for carry information or action to Ledger
//validated transaction will be added to block and updates state correspondingly

var TxStore ILedgerStore

type Transaction struct {
	TxType         TransactionType
	PayloadVersion byte
	Payload        Payload
	Attributes     []*TxAttribute
	UTXOInputs     []*UTXOTxInput
	BalanceInputs  []*BalanceTxInput
	Outputs        []*TxOutput
	SystemFee      Fixed64

	Programs []*program.Program

	//cache only, needn't serialize
	referTx    []*TxOutput
	hash       *Uint256
	networkFee Fixed64
}

//Serialize the Transaction
func (tx *Transaction) Serialize(w io.Writer) error {

	err := tx.SerializeUnsigned(w)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction txSerializeUnsigned Serialize failed.")
	}
	//Serialize  Transaction's programs
	lens := uint64(len(tx.Programs))
	err = serialization.WriteVarUint(w, lens)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction WriteVarUint failed.")
	}
	if lens > 0 {
		for _, p := range tx.Programs {
			err = p.Serialize(w)
			if err != nil {
				return NewDetailErr(err, ErrNoCode, "Transaction Programs Serialize failed.")
			}
		}
	}
	return nil
}

//Serialize the Transaction data without contracts
func (tx *Transaction) SerializeUnsigned(w io.Writer) error {
	//txType
	w.Write([]byte{byte(tx.TxType)})
	//PayloadVersion
	w.Write([]byte{tx.PayloadVersion})
	//Payload
	if tx.Payload == nil {
		return errors.New("Transaction Payload is nil.")
	}
	tx.Payload.Serialize(w, tx.PayloadVersion)
	//[]*txAttribute
	err := serialization.WriteVarUint(w, uint64(len(tx.Attributes)))
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction item txAttribute length serialization failed.")
	}
	if len(tx.Attributes) > 0 {
		for _, attr := range tx.Attributes {
			attr.Serialize(w)
		}
	}
	//[]*UTXOInputs
	err = serialization.WriteVarUint(w, uint64(len(tx.UTXOInputs)))
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction item UTXOInputs length serialization failed.")
	}
	if len(tx.UTXOInputs) > 0 {
		for _, utxo := range tx.UTXOInputs {
			utxo.Serialize(w)
		}
	}
	// TODO BalanceInputs
	//[]*Outputs
	err = serialization.WriteVarUint(w, uint64(len(tx.Outputs)))
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction item Outputs length serialization failed.")
	}
	if len(tx.Outputs) > 0 {
		for _, output := range tx.Outputs {
			output.Serialize(w)
		}
	}
	err = tx.SystemFee.Serialize(w)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction item SystemFee serialization failed.")
	}
	return nil
}

//deserialize the Transaction
func (tx *Transaction) Deserialize(r io.Reader) error {
	// tx deserialize
	err := tx.DeserializeUnsigned(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "transaction Deserialize error")
	}
	// tx networkFee
	tx.networkFee = -1
	// tx program
	lens, err := serialization.ReadVarUint(r, 0)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "transaction tx program Deserialize error")
	}

	programHashes := []*program.Program{}
	if lens > 0 {
		for i := 0; i < int(lens); i++ {
			outputHashes := new(program.Program)
			err := outputHashes.Deserialize(r)
			if err != nil {
				return errors.New("deserialize transaction failed")
			}
			programHashes = append(programHashes, outputHashes)
		}
		tx.Programs = programHashes
	}
	return nil
}

func (tx *Transaction) DeserializeUnsigned(r io.Reader) error {
	var txType [1]byte
	_, err := io.ReadFull(r, txType[:])
	if err != nil {
		return err
	}
	tx.TxType = TransactionType(txType[0])
	return tx.DeserializeUnsignedWithoutType(r)
}

func (tx *Transaction) DeserializeUnsignedWithoutType(r io.Reader) error {
	var payloadVersion [1]byte
	_, err := io.ReadFull(r, payloadVersion[:])
	tx.PayloadVersion = payloadVersion[0]
	if err != nil {
		return err
	}

	//payload
	//tx.Payload.Deserialize(r)
	switch tx.TxType {
	case RegisterAsset:
		tx.Payload = new(payload.RegisterAsset)
	case IssueAsset:
		tx.Payload = new(payload.IssueAsset)
	case TransferAsset:
		tx.Payload = new(payload.TransferAsset)
	case BookKeeping:
		tx.Payload = new(payload.BookKeeping)
	case Record:
		tx.Payload = new(payload.Record)
	case BookKeeper:
		tx.Payload = new(payload.BookKeeper)
	case PrivacyPayload:
		tx.Payload = new(payload.PrivacyPayload)
	case DataFile:
		tx.Payload = new(payload.DataFile)
	case Deploy:
		tx.Payload = new(payload.DeployCode)
	case Invoke:
		tx.Payload = new(payload.InvokeCode)
	case Claim:
		tx.Payload = new(payload.Claim)
	case Enrollment:
		tx.Payload = new(payload.Enrollment)
	case Vote:
		tx.Payload = new(payload.Vote)
	default:
		return errors.New("[Transaction],invalide transaction type.")
	}
	err = tx.Payload.Deserialize(r, tx.PayloadVersion)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Payload Parse error")
	}
	//attributes
	Len, err := serialization.ReadVarUint(r, 0)
	if err != nil {
		return err
	}
	if Len > uint64(0) {
		for i := uint64(0); i < Len; i++ {
			attr := new(TxAttribute)
			err = attr.Deserialize(r)
			if err != nil {
				return err
			}
			tx.Attributes = append(tx.Attributes, attr)
		}
	}
	//UTXOInputs
	Len, err = serialization.ReadVarUint(r, 0)
	if err != nil {
		return err
	}
	if Len > uint64(0) {
		for i := uint64(0); i < Len; i++ {
			utxo := new(UTXOTxInput)
			err = utxo.Deserialize(r)
			if err != nil {
				return err
			}
			tx.UTXOInputs = append(tx.UTXOInputs, utxo)
		}
	}
	//TODO balanceInputs
	//Outputs
	Len, err = serialization.ReadVarUint(r, 0)
	if err != nil {
		return err
	}
	if Len > uint64(0) {
		for i := uint64(0); i < Len; i++ {
			output := new(TxOutput)
			output.Deserialize(r)

			tx.Outputs = append(tx.Outputs, output)
		}
	}
	err = tx.SystemFee.Deserialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction item SystemFee Deserialize failed.")
	}
	return nil
}

func (tx *Transaction) GetProgramHashes() ([]Uint160, error) {
	if tx == nil {
		return []Uint160{}, errors.New("[Transaction],GetProgramHashes transaction is nil.")
	}
	hashs := []Uint160{}
	uniqHashes := []Uint160{}
	// add inputUTXO's transaction
	referOutput, err := tx.GetReference()
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "[Transaction], GetProgramHashes failed.")
	}
	for _, output := range referOutput {
		programHash := output.ProgramHash
		hashs = append(hashs, programHash)
	}
	for _, attribute := range tx.Attributes {
		if attribute.Usage == Script {
			dataHash, err := Uint160ParseFromBytes(attribute.Data)
			if err != nil {
				return nil, NewDetailErr(errors.New("[Transaction], GetProgramHashes err."), ErrNoCode, "")
			}
			hashs = append(hashs, Uint160(dataHash))
		}
	}
	switch tx.TxType {
	case RegisterAsset:
		issuer := tx.Payload.(*payload.RegisterAsset).Issuer
		signatureRedeemScript, err := contract.CreateSignatureRedeemScript(issuer)
		if err != nil {
			return nil, NewDetailErr(err, ErrNoCode, "[Transaction], GetProgramHashes CreateSignatureRedeemScript failed.")
		}

		astHash := ToCodeHash(signatureRedeemScript)
		hashs = append(hashs, astHash)
	case IssueAsset:
		result := tx.GetMergedAssetIDValueFromOutputs()
		if err != nil {
			return nil, NewDetailErr(err, ErrNoCode, "[Transaction], GetTransactionResults failed.")
		}
		for k := range result {
			if k.CompareTo(ONTTokenID) == 0 || k.CompareTo(ONGTokenID) == 0 {
				continue
			}
			tx, err := TxStore.GetTransaction(k)
			if err != nil {
				return nil, NewDetailErr(err, ErrNoCode, fmt.Sprintf("[Transaction], GetTransaction failed With AssetID:=%x", k))
			}
			if tx.TxType != RegisterAsset {
				return nil, NewDetailErr(errors.New("[Transaction] error"), ErrNoCode, fmt.Sprintf("[Transaction], Transaction Type ileage With AssetID:=%x", k))
			}

			switch v1 := tx.Payload.(type) {
			case *payload.RegisterAsset:
				hashs = append(hashs, v1.Controller)
			default:
				return nil, NewDetailErr(errors.New("[Transaction] error"), ErrNoCode, fmt.Sprintf("[Transaction], payload is illegal", k))
			}
		}
	case DataFile:
		issuer := tx.Payload.(*payload.DataFile).Issuer
		signatureRedeemScript, err := contract.CreateSignatureRedeemScript(issuer)
		if err != nil {
			return nil, NewDetailErr(err, ErrNoCode, "[Transaction], GetProgramHashes CreateSignatureRedeemScript failed.")
		}

		astHash := ToCodeHash(signatureRedeemScript)
		hashs = append(hashs, astHash)
	case TransferAsset:
	case Record:
	case BookKeeper:
		issuer := tx.Payload.(*payload.BookKeeper).Issuer
		signatureRedeemScript, err := contract.CreateSignatureRedeemScript(issuer)
		if err != nil {
			return nil, NewDetailErr(err, ErrNoCode, "[Transaction - BookKeeper], GetProgramHashes CreateSignatureRedeemScript failed.")
		}

		astHash := ToCodeHash(signatureRedeemScript)
		hashs = append(hashs, astHash)
	case PrivacyPayload:
		issuer := tx.Payload.(*payload.PrivacyPayload).EncryptAttr.(*payload.EcdhAes256).FromPubkey
		signatureRedeemScript, err := contract.CreateSignatureRedeemScript(issuer)
		if err != nil {
			return nil, NewDetailErr(err, ErrNoCode, "[Transaction], GetProgramHashes CreateSignatureRedeemScript failed.")
		}

		astHash := ToCodeHash(signatureRedeemScript)
		hashs = append(hashs, astHash)
	case Claim:
		// add claim UTXO's in to check list
		reference := make(map[*UTXOTxInput]*TxOutput)
		// Key index，v UTXOInput
		for _, utxo := range tx.Payload.(*payload.Claim).Claims {
			transaction, err := TxStore.GetTransaction(utxo.ReferTxID)
			if err != nil {
				return nil, NewDetailErr(err, ErrNoCode, "[Transaction], GetReference failed.")
			}
			index := utxo.ReferTxOutputIndex
			reference[utxo] = transaction.Outputs[index]
		}
		for _, output := range reference {
			programHash := output.ProgramHash
			hashs = append(hashs, programHash)
		}
	case Vote:
		vote := tx.Payload.(*payload.Vote)
		hash := vote.Account
		hashs = append(hashs, hash)
	default:
	}
	//remove dupilicated hashes
	uniq := make(map[Uint160]bool)
	for _, v := range hashs {
		uniq[v] = true
	}
	for k := range uniq {
		uniqHashes = append(uniqHashes, k)
	}
	sort.Sort(byProgramHashes(uniqHashes))
	return uniqHashes, nil
}

func (tx *Transaction) SetPrograms(programs []*program.Program) {
	tx.Programs = programs
}

func (tx *Transaction) GetPrograms() []*program.Program {
	return tx.Programs
}

func (tx *Transaction) GetOutputHashes() ([]Uint160, error) {
	//TODO: implement Transaction.GetOutputHashes()

	return []Uint160{}, nil
}

func (tx *Transaction) GetMessage() []byte {
	return sig.GetHashData(tx)
}

func (tx *Transaction) ToArray() []byte {
	b := new(bytes.Buffer)
	tx.Serialize(b)
	return b.Bytes()
}

func (tx *Transaction) Hash() Uint256 {
	if tx.hash == nil {
		d := sig.GetHashData(tx)
		temp := sha256.Sum256([]byte(d))
		f := Uint256(sha256.Sum256(temp[:]))
		tx.hash = &f
	}
	return *tx.hash

}

func (tx *Transaction) SetHash(hash Uint256) {
	tx.hash = &hash
}

func (tx *Transaction) Type() InventoryType {
	return TRANSACTION
}
func (tx *Transaction) Verify() error {
	//TODO: Verify()
	return nil
}

func (tx *Transaction) GetReference() ([]*TxOutput, error) {
	if tx.referTx != nil {
		return tx.referTx, nil
	}
	if len(tx.UTXOInputs) <= 0 {
		tx.referTx = []*TxOutput{}
		return tx.referTx, nil
	}
	//UTXO input /  Outputs
	reference := make([]*TxOutput, 0, len(tx.UTXOInputs))
	// Key index，v UTXOInput
	for _, utxo := range tx.UTXOInputs {
		transaction, err := TxStore.GetTransaction(utxo.ReferTxID)
		if err != nil {
			return nil, NewDetailErr(err, ErrNoCode, "[Transaction], GetReference failed.")
		}
		index := utxo.ReferTxOutputIndex
		reference = append(reference, transaction.Outputs[index])
	}
	tx.referTx = reference
	return reference, nil
}

func (tx *Transaction) GetTransactionResults() (TransactionResult, error) {
	result := make(map[Uint256]Fixed64)
	outputResult := tx.GetMergedAssetIDValueFromOutputs()
	InputResult, err := tx.GetMergedAssetIDValueFromReference()
	if err != nil {
		return nil, err
	}
	//calc the balance of input vs output
	for outputAssetid, outputValue := range outputResult {
		if inputValue, ok := InputResult[outputAssetid]; ok {
			result[outputAssetid] = inputValue - outputValue
		} else {
			result[outputAssetid] -= outputValue
		}
	}
	for inputAssetid, inputValue := range InputResult {
		if _, exist := result[inputAssetid]; !exist {
			result[inputAssetid] += inputValue
		}
	}
	return result, nil
}

func (tx *Transaction) GetMergedAssetIDValueFromOutputs() TransactionResult {
	var result = make(map[Uint256]Fixed64)
	for _, v := range tx.Outputs {
		amout, ok := result[v.AssetID]
		if ok {
			result[v.AssetID] = amout + v.Value
		} else {
			result[v.AssetID] = v.Value
		}
	}
	return result
}

func (tx *Transaction) GetMergedAssetIDValueFromReference() (TransactionResult, error) {
	reference, err := tx.GetReference()
	if err != nil {
		return nil, err
	}
	var result = make(map[Uint256]Fixed64)
	for _, v := range reference {
		amout, ok := result[v.AssetID]
		if ok {
			result[v.AssetID] = amout + v.Value
		} else {
			result[v.AssetID] = v.Value
		}
	}
	return result, nil
}

func (tx *Transaction) GetSysFee() Fixed64 {
	return Fixed64(Parameters.SystemFee[TxName[tx.TxType]])
}

func (tx *Transaction) GetNetworkFee() (Fixed64, error) {
	if tx.networkFee != -1 {
		return tx.networkFee, nil
	}
	txHash := tx.Hash()
	if txHash.CompareTo(SystemIssue) == 0 || tx.TxType == Claim || tx.TxType == BookKeeping {
		return 0, nil
	}
	refrence, err := tx.GetReference()
	if err != nil {
		return 0, errors.New(fmt.Sprintf("[GetNetworkFee], GetRefrence error：%v", err))
	}
	var input int64
	for _, v := range refrence {
		if v.AssetID.CompareTo(ONGTokenID) == 0 {
			input += v.Value.GetData()
		}
	}
	var output int64
	for _, v := range tx.Outputs {
		if v.AssetID.CompareTo(ONGTokenID) == 0 {
			output += v.Value.GetData()
		}
	}
	result := Fixed64(input - output - tx.SystemFee.GetData())
	if result >= 0 {
		tx.networkFee = result
		return result, nil
	} else {
		return 0, errors.New("[GetNetworkFee] failed as invalid network fee.")
	}
}

func NewGoverningToken() *Transaction {
	regAsset, _ := NewRegisterAssetTransaction(
		&asset.Asset{
			Name:        "ONT",
			Description: "Ontology Network ONT Token",
			Precision:   0,
			AssetType:   asset.GoverningToken,
			RecordType:  asset.UTXO,
		},
		FromDecimal(OngRegisterAmount),
		Infinity,
		ToCodeHash([]byte{byte(vm.PUSHF)}),
	)
	return regAsset
}

func NewUtilityToken() *Transaction {
	regAsset, _ := NewRegisterAssetTransaction(
		&asset.Asset{
			Name:        "ONG",
			Description: "Ontology Network ONG Token",
			Precision:   8,
			AssetType:   asset.UtilityToken,
			RecordType:  asset.UTXO,
		},
		FromDecimal(OngRegisterAmount),
		Infinity,
		ToCodeHash([]byte{byte(vm.PUSHF)}),
	)
	return regAsset
}

func NewIssueToken(governingToken, utilityToken *Transaction) *Transaction {
	//bookKeepers := crypto.GetBookKeepers()
	//multiSigContract, _ := contract.CreateMultiSigRedeemScript(len(bookKeepers) / 2 + 1, bookKeepers)
	programHexStr := `de16a89b7fed8974ea635867b23ffed6ea53ef51`
	programByte, _ := HexToBytes(programHexStr)
	programHash, _ := Uint160ParseFromBytes(programByte)
	issueAsset, _ := NewIssueAssetTransaction(
		[]*TxOutput{
			&TxOutput{
				AssetID:     governingToken.Hash(),
				Value:       governingToken.Payload.(*payload.RegisterAsset).Amount,
				ProgramHash: programHash,
			},
			&TxOutput{
				AssetID:     utilityToken.Hash(),
				Value:       Fixed64(governingToken.Payload.(*payload.RegisterAsset).Amount.GetData() * 10 / 100),
				ProgramHash: programHash,
			},
		})
	SystemIssue = issueAsset.Hash()
	return issueAsset
}

type byProgramHashes []Uint160

func (a byProgramHashes) Len() int {
	return len(a)
}
func (a byProgramHashes) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a byProgramHashes) Less(i, j int) bool {
	if a[i].CompareTo(a[j]) > 0 {
		return false
	} else {
		return true
	}
}

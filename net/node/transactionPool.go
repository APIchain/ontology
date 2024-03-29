package node

import (
	"fmt"
	"github.com/Ontology/common"
	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/ledger"
	"github.com/Ontology/core/transaction"
	"github.com/Ontology/core/transaction/payload"
	"github.com/Ontology/core/transaction/utxo"
	va "github.com/Ontology/core/validation"
	ontError "github.com/Ontology/errors"
	"sort"
	"sync"
)

// Genesis transaction will not be added to pool, so use height == 0 to indicate transaction not packed in block
type PoolTransaction struct {
	tx *transaction.Transaction
	//height uint32
	fee common.Fixed64
}

type TXNPool struct {
	sync.RWMutex
	txnList       map[common.Uint256]*PoolTransaction // transaction which have been verifyed will put into this map
	issueSummary  map[common.Uint256]common.Fixed64   // transaction which pass the verify will summary the amout to this map
	inputUTXOList map[string]*transaction.Transaction // transaction which pass the verify will add the UTXO to this map
}

func (this *TXNPool) init() {
	this.inputUTXOList = make(map[string]*transaction.Transaction)
	this.issueSummary = make(map[common.Uint256]common.Fixed64)
	this.txnList = make(map[common.Uint256]*PoolTransaction)
}

//append transaction to txnpool when check ok.
//1.check transaction. 2.check with ledger(db) 3.check with pool
func (this *TXNPool) AppendTxnPool(txn *transaction.Transaction) ontError.ErrCode {
	//verify transaction with Concurrency
	if errCode := va.VerifyTransaction(txn); errCode != ontError.ErrNoError {
		log.Infof("Transaction verification failed %x\n", txn.Hash())
		return errCode
	}
	if errCode := va.VerifyTransactionWithLedger(txn, ledger.DefaultLedger); errCode != ontError.ErrNoError {
		log.Infof("Transaction verification with ledger failed %x\n", txn.Hash())
		return errCode
	}
	//verify transaction by pool with lock
	if errCode := this.verifyTransactionWithTxnPool(txn); errCode != ontError.ErrNoError {
		log.Info("Transaction verification with transaction pool failed", txn.Hash())
		return errCode
	}
	//add the transaction to process scope
	this.addtxnList(txn)
	return ontError.ErrNoError
}

//get the transaction in txnpool
func (this *TXNPool) GetTxnPool(byCount bool) (map[common.Uint256]*transaction.Transaction, common.Fixed64) {
	this.Lock()

	orderByFee := make([]*PoolTransaction, 0, len(this.txnList))
	for _, ptx := range this.txnList {
		orderByFee = append(orderByFee, ptx)
	}
	sort.Sort(OrderByNetWorkFee(orderByFee))

	maxcount := config.Parameters.MaxTxInBlock
	if maxcount <= 0 || byCount == false || maxcount > len(orderByFee) {
		maxcount = len(orderByFee)
	}

	txnMap := make(map[common.Uint256]*transaction.Transaction, maxcount)
	var networkFeeSum common.Fixed64
	for i := 0; i < maxcount; i++ {
		ptx := orderByFee[i]
		networkFeeSum += common.Fixed64(ptx.fee)
		txnMap[ptx.tx.Hash()] = ptx.tx
	}

	this.Unlock()
	return txnMap, networkFeeSum
}

//get the transaction in txnpool
func (this *TXNPool) GetTxnPoolTxlist() []common.Uint256 {
	this.RLock()
	list := make([]common.Uint256, 0, len(this.txnList))
	for k, _ := range this.txnList {
		list = append(list, k)
	}
	this.RUnlock()
	return list
}

//clean the trasaction Pool with committed block.
func (this *TXNPool) CleanSubmittedTransactions(block *ledger.Block) error {
	this.cleanTransactionList(block.Transactions)
	this.cleanUTXOList(block.Transactions)
	this.cleanIssueSummary(block.Transactions)
	return nil
}

//get the transaction by hash
func (this *TXNPool) GetTransaction(hash common.Uint256) *transaction.Transaction {
	this.RLock()
	defer this.RUnlock()
	return this.txnList[hash].tx
}

//verify transaction with txnpool
func (this *TXNPool) verifyTransactionWithTxnPool(txn *transaction.Transaction) ontError.ErrCode {
	// check if the transaction includes double spent UTXO inputs
	ok, duplicateTxn := this.apendToUTXOPool(txn)
	if !ok && duplicateTxn != nil {
		log.Info(fmt.Sprintf("txn=%x duplicateTxn UTXO occurs with txn in pool=%x,keep the latest one.", txn.Hash(), duplicateTxn.Hash()))
		this.removeTransaction(duplicateTxn)
	}
	//check issue transaction weather occur exceed issue range.
	if ok := this.summaryAssetIssueAmount(txn); !ok {
		log.Info(fmt.Sprintf("Check summary Asset Issue Amount failed with txn=%x", txn.Hash()))
		this.removeTransaction(txn)
		return ontError.ErrSummaryAsset
	}

	return ontError.ErrNoError
}

//remove from associated map
func (this *TXNPool) removeTransaction(txn *transaction.Transaction) {
	//1.remove from txnList
	this.deltxnList(txn)
	//2.remove from UTXO list map
	for _, input := range txn.UTXOInputs {
		this.delInputUTXOList(input)
	}
	//3.remove From Asset Issue Summary map
	if txn.TxType != transaction.IssueAsset {
		return
	}
	transactionResult := txn.GetMergedAssetIDValueFromOutputs()
	for k, delta := range transactionResult {
		this.decrAssetIssueAmountSummary(k, delta)
	}
}

//check and add to utxo list pool
func (this *TXNPool) apendToUTXOPool(txn *transaction.Transaction) (bool, *transaction.Transaction) {
	for _, input := range txn.UTXOInputs {
		t := this.getInputUTXOList(input)
		if t != nil {
			return false, t
		}
		this.addInputUTXOList(txn, input)
	}
	return true, nil
}

//clean txnpool utxo map
func (this *TXNPool) cleanUTXOList(txs []*transaction.Transaction) {
	for _, txn := range txs {
		for _, input := range txn.UTXOInputs {
			this.delInputUTXOList(input)
		}
	}
}

//check and summary to issue amount Pool
func (this *TXNPool) summaryAssetIssueAmount(txn *transaction.Transaction) bool {
	if txn.TxType != transaction.IssueAsset {
		return true
	}
	transactionResult := txn.GetMergedAssetIDValueFromOutputs()
	for k, delta := range transactionResult {
		//Check weather occur exceed the amount when RegisterAsseted
		//1. Get the Asset amount when RegisterAsseted.
		txn, err := transaction.TxStore.GetTransaction(k)
		if err != nil {
			return false
		}
		if txn.TxType != transaction.RegisterAsset {
			return false
		}
		AssetReg := txn.Payload.(*payload.RegisterAsset)

		//2. Get the amount has been issued of this assetID
		var quantity_issued common.Fixed64
		if AssetReg.Amount < common.Fixed64(0) {
			continue
		} else {
			quantity_issued, err = transaction.TxStore.GetQuantityIssued(k)
			if err != nil {
				return false
			}
		}

		//3. calc weather out off the amount when Registed.
		//AssetReg.Amount : amount when RegisterAsset of this assedID
		//quantity_issued : amount has been issued of this assedID
		//txnPool.issueSummary[k] : amount in transactionPool of this assedID
		if AssetReg.Amount < quantity_issued+this.getAssetIssueAmount(k)+delta {
			return false
		}
	}

	for k, delta := range transactionResult {
		//update the amount in txnPool
		this.incrAssetIssueAmountSummary(k, delta)
	}

	return true
}

// clean the trasaction Pool with committed transactions.
func (this *TXNPool) cleanTransactionList(txns []*transaction.Transaction) error {
	cleaned := 0
	txnsNum := len(txns)
	for _, txn := range txns {
		if txn.TxType == transaction.BookKeeping {
			txnsNum = txnsNum - 1
			continue
		}
		if this.deltxnList(txn) {
			cleaned++
		}
	}
	if txnsNum != cleaned {
		log.Info(fmt.Sprintf("The Transactions num Unmatched. Expect %d, got %d .\n", txnsNum, cleaned))
	}
	log.Debug(fmt.Sprintf("[cleanTransactionList],transaction %d Requested, %d cleaned, Remains %d in TxPool", txnsNum, cleaned, this.GetTransactionCount()))
	return nil
}

func (this *TXNPool) addtxnList(txn *transaction.Transaction) bool {
	this.Lock()
	defer this.Unlock()
	txnHash := txn.Hash()
	if _, ok := this.txnList[txnHash]; ok {
		return false
	}
	nfw, err := txn.GetNetworkFee()
	if err != nil {
		return false
	}
	this.txnList[txnHash] = &PoolTransaction{tx: txn, fee: nfw}
	return true
}

func (this *TXNPool) deltxnList(tx *transaction.Transaction) bool {
	this.Lock()
	defer this.Unlock()
	txHash := tx.Hash()
	if _, ok := this.txnList[txHash]; !ok {
		return false
	}
	delete(this.txnList, txHash)
	return true
}

func (this *TXNPool) copytxnList() map[common.Uint256]*transaction.Transaction {
	this.RLock()
	defer this.RUnlock()
	txnMap := make(map[common.Uint256]*transaction.Transaction, len(this.txnList))
	for txnId, txn := range this.txnList {
		txnMap[txnId] = txn.tx
	}
	return txnMap
}

func (this *TXNPool) GetTransactionCount() int {
	this.RLock()
	defer this.RUnlock()
	return len(this.txnList)
}

func (this *TXNPool) getInputUTXOList(input *utxo.UTXOTxInput) *transaction.Transaction {
	this.RLock()
	defer this.RUnlock()
	return this.inputUTXOList[input.ToString()]
}

func (this *TXNPool) addInputUTXOList(tx *transaction.Transaction, input *utxo.UTXOTxInput) bool {
	this.Lock()
	defer this.Unlock()
	id := input.ToString()
	_, ok := this.inputUTXOList[id]
	if ok {
		return false
	}
	this.inputUTXOList[id] = tx

	return true
}

func (this *TXNPool) delInputUTXOList(input *utxo.UTXOTxInput) bool {
	this.Lock()
	defer this.Unlock()
	id := input.ToString()
	_, ok := this.inputUTXOList[id]
	if !ok {
		return false
	}
	delete(this.inputUTXOList, id)
	return true
}

func (this *TXNPool) incrAssetIssueAmountSummary(assetId common.Uint256, delta common.Fixed64) {
	this.Lock()
	defer this.Unlock()
	this.issueSummary[assetId] = this.issueSummary[assetId] + delta
}

func (this *TXNPool) decrAssetIssueAmountSummary(assetId common.Uint256, delta common.Fixed64) {
	this.Lock()
	defer this.Unlock()
	amount, ok := this.issueSummary[assetId]
	if !ok {
		return
	}
	amount = amount - delta
	if amount < common.Fixed64(0) {
		amount = common.Fixed64(0)
	}
	this.issueSummary[assetId] = amount
}

func (this *TXNPool) cleanIssueSummary(txs []*transaction.Transaction) {
	for _, v := range txs {
		if v.TxType == transaction.IssueAsset {
			transactionResult := v.GetMergedAssetIDValueFromOutputs()
			for k, delta := range transactionResult {
				this.decrAssetIssueAmountSummary(k, delta)
			}
		}
	}
}

func (this *TXNPool) getAssetIssueAmount(assetId common.Uint256) common.Fixed64 {
	this.RLock()
	defer this.RUnlock()
	return this.issueSummary[assetId]
}

type OrderByNetWorkFee []*PoolTransaction

func (n OrderByNetWorkFee) Len() int { return len(n) }

func (n OrderByNetWorkFee) Swap(i, j int) { n[i], n[j] = n[j], n[i] }

func (n OrderByNetWorkFee) Less(i, j int) bool { return n[j].fee < n[i].fee }

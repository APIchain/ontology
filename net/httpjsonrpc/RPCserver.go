package httpjsonrpc

import (
	. "github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	"net/http"
	"strconv"
)

func StartRPCServer() {
	log.Debug()
	http.HandleFunc("/", Handle)

	HandleFunc("getbestblockhash", getBestBlockHash)
	HandleFunc("getblock", getBlock)
	HandleFunc("getblockcount", getBlockCount)
	HandleFunc("getblockhash", getBlockHash)
	HandleFunc("getunspendoutput", getUnspendOutput)
	HandleFunc("getconnectioncount", getConnectionCount)
	HandleFunc("getrawmempool", getRawMemPool)
	HandleFunc("getrawtransaction", getRawTransaction)
	HandleFunc("getcalculateBouns", getCalculateBouns)
	HandleFunc("sendrawtransaction", sendRawTransaction)
	HandleFunc("getstorage", getStorage)
	HandleFunc("getbalance", getBalance)
	HandleFunc("submitblock", submitBlock)
	HandleFunc("getversion", getVersion)
	HandleFunc("getdataile", getDataFile)
	HandleFunc("catdatarecord", catDataRecord)
	HandleFunc("regdatafile", regDataFile)
	HandleFunc("uploadDataFile", uploadDataFile)
	HandleFunc("getsmartcodeevent", getSmartCodeEvent)

	err := http.ListenAndServe(":" + strconv.Itoa(Parameters.HttpJsonPort), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}

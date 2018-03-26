/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package genesis

import (
	"errors"
	"fmt"
	"time"

	"github.com/Ontology/common"
	"github.com/Ontology/common/config"
	vconfig "github.com/Ontology/consensus/vbft/config"
	"github.com/Ontology/core/types"
	"github.com/Ontology/core/utils"
	"github.com/Ontology/crypto"
	vmtypes "github.com/Ontology/vm/types"
	"github.com/Ontology/smartcontract/service/native/states"
	"bytes"
)

const (
	BlockVersion      uint32 = 0
	GenesisNonce      uint64 = 2083236893

	OntRegisterAmount = 1000000000
	OngRegisterAmount = 1000000000
)

var (
	OntContractAddress, _ = common.AddressParseFromBytes([]byte{0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01})
	OngContractAddress, _ = common.AddressParseFromBytes([]byte{0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02})

	ONTToken   = NewGoverningToken()
	ONGToken   = NewUtilityToken()
	ONTTokenID = ONTToken.Hash()
	ONGTokenID = ONGToken.Hash()
)

var GenBlockTime = (config.DEFAULTGENBLOCKTIME * time.Second)

var GenesisBookkeepers []*crypto.PubKey

func GenesisBlockInit(defaultBookkeeper []*crypto.PubKey) (*types.Block, error) {
	//getBookkeeper
	GenesisBookkeepers = defaultBookkeeper
	nextBookkeeper, err := types.AddressFromBookkeepers(defaultBookkeeper)
	if err != nil {
		return nil, errors.New("[Block],GenesisBlockInit err with GetBookkeeperAddress")
	}

	consensusPayload, err := vconfig.GenesisConsensusPayload()
	if err != nil {
		return nil, fmt.Errorf("consensus genesus init failed: %s", err)
	}

	//blockdata
	genesisHeader := &types.Header{
		Version:          BlockVersion,
		PrevBlockHash:    common.Uint256{},
		TransactionsRoot: common.Uint256{},
		Timestamp:        uint32(uint32(time.Date(2017, time.February, 23, 0, 0, 0, 0, time.UTC).Unix())),
		Height:           uint32(0),
		ConsensusData:    GenesisNonce,
		ConsensusPayload: consensusPayload,
		NextBookkeeper:   nextBookkeeper,

		Bookkeepers: nil,
		SigData:     nil,
	}

	//block
	ont := NewGoverningToken()
	ong := NewUtilityToken()

	genesisBlock := &types.Block{
		Header: genesisHeader,
		Transactions: []*types.Transaction{
			ont,
			ong,
			NewGoverningInit(),
			NewUtilityInit(),
		},
	}
	return genesisBlock, nil
}

func NewGoverningToken() *types.Transaction {
	tx := utils.NewDeployTransaction(&vmtypes.VmCode{Code: OntContractAddress[:], VmType: vmtypes.Native}, "ONT", "1.0",
		"Ontology Team", "contact@ont.io", "Ontology Network ONT Token", true)
	return tx
}

func NewUtilityToken() *types.Transaction {
	tx := utils.NewDeployTransaction(&vmtypes.VmCode{Code: OngContractAddress[:], VmType: vmtypes.Native}, "ONG", "1.0",
		"Ontology Team", "contact@ont.io", "Ontology Network ONG Token", true)
	return tx
}

func NewGoverningInit() *types.Transaction {
	init := states.Contract{
		Address: OntContractAddress,
		Method: "init",
		Args: []byte{},
	}
	bf := new(bytes.Buffer)
	init.Serialize(bf)
	vmCode := vmtypes.VmCode{
		VmType: vmtypes.Native,
		Code: bf.Bytes(),
	}
	tx := utils.NewInvokeTransaction(vmCode)
	return tx
}

func NewUtilityInit() *types.Transaction {
	init := states.Contract{
		Address: OngContractAddress,
		Method: "init",
		Args: []byte{},
	}
	bf := new(bytes.Buffer)
	init.Serialize(bf)
	vmCode := vmtypes.VmCode{
		VmType: vmtypes.Native,
		Code: bf.Bytes(),
	}
	tx := utils.NewInvokeTransaction(vmCode)
	return tx
}

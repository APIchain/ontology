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

package net

import (
	ns "github.com/Ontology/net/actor"
	"github.com/Ontology/net/node"
	"github.com/Ontology/net/protocol"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-eventbus/actor"
)

func SetTxnPoolPid(txnPid *actor.PID) {
	ns.SetTxnPoolPid(txnPid)
}

func SetConsensusPid(conPid *actor.PID) {
	ns.SetConsensusPid(conPid)
}

func SetLedgerPid(conPid *actor.PID) {
	ns.SetLedgerPid(conPid)
}

func InitNetServerActor(noder protocol.Noder) (*actor.PID, error) {
	netServerPid, err := ns.InitNetServer(noder)
	return netServerPid, err
}

func StartProtocol(pubKey keypair.PublicKey) protocol.Noder {
	net := node.InitNode(pubKey)
	net.ConnectSeeds()
	return net
}

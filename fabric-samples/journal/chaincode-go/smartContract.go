/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/fwoie/fabric-journal/fabric-samples/journal/chaincode-go/smart-contract"
)

func main() {
	journalSmartContract, err := contractapi.NewChaincode(&journal.SmartContract{})
	if err != nil {
		log.Panicf("Error creating journal chaincode: %v", err)
	}

	if err := journalSmartContract.Start(); err != nil {
		log.Panicf("Error starting journal chaincode: %v", err)
	}
}

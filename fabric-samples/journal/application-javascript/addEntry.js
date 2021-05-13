/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { Gateway, Wallets } = require('fabric-network');
const path = require('path');
const { buildCCPOrg1, buildCCPOrg2, buildWallet } = require('../../test-application/javascript/AppUtil.js');

const myChannel = 'mychannel';
const myChaincodeName = 'journal';


function prettyJSONString(inputString) {
    if (inputString) {
        return JSON.stringify(JSON.parse(inputString), null, 2);
    }
    else {
        return inputString;
    }
}

async function addEntry(ccp,wallet,user,journalID,entryID,data) {
    try {

        const gateway = new Gateway();
      //connect using Discovery enabled

      await gateway.connect(ccp,
          { wallet: wallet, identity: user, discovery: { enabled: true, asLocalhost: true } });

        const network = await gateway.getNetwork(myChannel);
        const contract = network.getContract(myChaincodeName);

        let statefulTxn = contract.createTransaction('AddEntry');

        console.log('\n--> Submit Transaction: Add entry to journal');
        await statefulTxn.submit(journalID, entryID, data);
        console.log('*** Result: committed');

        console.log('\n--> Evaluate Transaction: query entry that was just created');
        let result = await contract.evaluateTransaction('GetEntry',journalID, entryID);
        console.log('*** Result: Journal: ' + prettyJSONString(result.toString()));

        gateway.disconnect();
    } catch (error) {
        console.error(`******** FAILED to add entry to journal: ${error}`);
	}
}

async function main() {
    try {

        if (process.argv[2] == undefined || process.argv[3] == undefined
            || process.argv[4] == undefined || process.argv[5] == undefined || process.argv[6] == undefined) {
            console.log("Usage: node createJournal.js org userID journalID entryID data");
            process.exit(1);
        }

        const org = process.argv[2]
        const user = process.argv[3];
        const journalID = process.argv[4];
        const entryID = process.argv[5];
        const data = process.argv[6];

        if (org == 'Org1' || org == 'org1') {

            const orgMSP = 'Org1MSP';
            const ccp = buildCCPOrg1();
            const walletPath = path.join(__dirname, 'wallet/org1');
            const wallet = await buildWallet(Wallets, walletPath);
            await addEntry(ccp,wallet,user,journalID,entryID,data);
        }
        else if (org == 'Org2' || org == 'org2') {

            const orgMSP = 'Org2MSP';
            const ccp = buildCCPOrg2();
            const walletPath = path.join(__dirname, 'wallet/org2');
            const wallet = await buildWallet(Wallets, walletPath);
            await addEntry(ccp,wallet,user,journalID,entryID,data);
        }  else {
            console.log("Usage: node createJournal.js org userID journalID entryID data");
            console.log("Org must be Org1 or Org2");
          }
    } catch (error) {
		console.error(`******** FAILED to run the application: ${error}`);
    }
}


main();

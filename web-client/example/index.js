const {XMLHttpRequest} = require("xmlhttprequest");
global.XMLHttpRequest = XMLHttpRequest;

const {TX_HASH, TX_CHAIN_ID} = require('./variables.js');
const MinterHub = require('..');

const minterHub = new MinterHub('http://46.101.215.17:9091');

logPromise(minterHub.getTokenList());

logPromise(minterHub.getOraclePriceList());

logPromise(minterHub.getTxStatus(TX_CHAIN_ID, TX_HASH));

logPromise(minterHub.getOracleEthFee());

function logPromise(promise) {
    promise.then((result) => console.log(result))
        .catch((error) => {
            console.log(error);
            throw error;
        })
}


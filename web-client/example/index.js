const {XMLHttpRequest} = require("xmlhttprequest");
global.XMLHttpRequest = XMLHttpRequest;

const {TX_HASH, GRPC_APU_URL} = require('./variables.js');
const MinterHub = require('..');

const minterHub = new MinterHub(GRPC_APU_URL);

logPromise(minterHub.getTokenList());

logPromise(minterHub.getOraclePriceList());

logPromise(minterHub.getTxStatus(TX_HASH));

logPromise(minterHub.getOracleEthFee());

function logPromise(promise) {
    promise.then((result) => console.log(result))
        .catch((error) => {
            console.log(error);
            throw error;
        })
}


const {XMLHttpRequest} = require("xmlhttprequest");
global.XMLHttpRequest = XMLHttpRequest;

const {TX_HASH, GRPC_APU_URL, ADDRESS } = require('./variables.js');
const MinterHub = require('..');

const minterHub = new MinterHub(GRPC_APU_URL);

logPromise(minterHub.getTokenList());

logPromise(minterHub.getOraclePriceList());

logPromise(minterHub.getTxStatus(TX_HASH));

logPromise(minterHub.getDiscountForHolder(ADDRESS));

logPromise(minterHub.getOracleEthFee());

logPromise(minterHub.getOracleBscFee());

function logPromise(promise) {
    promise.then((result) => console.log(result))
        .catch((error) => {
            console.log(error);
            throw error;
        })
}


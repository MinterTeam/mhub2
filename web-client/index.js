const atob = require('./utils/node-atob.js');

const {TokenInfosRequest, TransactionStatusRequest} = require('./gen/mhub2/v1/query_pb.js')
const {QueryClient: HubService} = require('./gen/mhub2/v1/query_grpc_web_pb.js')

const {QueryPricesRequest, QueryEthFeeRequest} = require('./gen/oracle/v1/query_pb.js')
const {QueryClient: OracleService} = require('./gen/oracle/v1/query_grpc_web_pb.js')


function MinterHub(hostname) {
    var hubService = new HubService(hostname);
    var oracleService = new OracleService(hostname);

    /**
     * @return {Promise<Array<TokenInfo.AsObject>>}
     */
    this.getTokenList = function() {
        return new Promise((resolve, reject) => {
            hubService.tokenInfos(new TokenInfosRequest(), {}, function(err, response) {
                if (err) {
                    reject(err);
                } else {
                    const tokenList = response.toObject().list.tokenInfosList.map((item) => {
                        item.commission = atob(item.commission);
                        return item;
                    })
                    resolve(tokenList);
                }
            });
        });
    }

    /**
     * @param {string} chainId
     * @param {string} txHash
     * @return {Promise<TxStatus.AsObject>}
     */
    this.getTxStatus = function(chainId, txHash) {
        return new Promise((resolve, reject) => {
            hubService.transactionStatus(new TransactionStatusRequest([chainId, txHash]), {}, function(err, response) {
                if (err) {
                    reject(err);
                } else {
                    resolve(response.toObject().status);
                }
            });
        });
    }

    /**
     * @return {Promise<Price.AsObject>}
     */
    this.getOraclePriceList = function() {
        return new Promise((resolve, reject) => {
            oracleService.prices(new QueryPricesRequest(), {}, function(err, response) {
                if (err) {
                    reject(err);
                } else {
                    resolve(response.toObject().prices.listList);
                }
            });
        });
    }

    /**
     * @return {Promise<QueryEthFeeResponse.AsObject>}
     */
    this.getOracleEthFee = function() {
        return new Promise((resolve, reject) => {
            oracleService.ethFee(new QueryEthFeeRequest(), {}, function(err, response) {
                if (err) {
                    reject(err);
                } else {
                    resolve(response.toObject());
                }
            });
        });
    }
}

module.exports = MinterHub;

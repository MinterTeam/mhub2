const {XMLHttpRequest} = require("xmlhttprequest");
global.XMLHttpRequest = XMLHttpRequest;

const {TX_HASH, TX_CHAIN_ID} = require('./variables.js');

const {ParamsRequest, TokenInfosRequest, TransactionStatusRequest} = require('../gen/mhub2/v1/query_pb.js')
const {QueryClient: HubService} = require('../gen/mhub2/v1/query_grpc_web_pb.js')

const {QueryPricesRequest, QueryEthFeeRequest} = require('../gen/oracle/v1/query_pb.js')
const {QueryClient: OracleService} = require('../gen/oracle/v1/query_grpc_web_pb.js')


var hubService = new HubService('http://46.101.215.17:9091');
var oracleService = new OracleService('http://46.101.215.17:9091');


// hubService.params(new ParamsRequest(), {}, function(err, response) {
//   if (err) {
//     console.log(err);
//   } else {
//     console.log(response.toObject().params);
//   }
// });

hubService.tokenInfos(new TokenInfosRequest(), {}, function(err, response) {
  if (err) {
    console.log(err);
  } else {
    console.log(response.toObject().list.tokenInfosList.map((item) => {
      item.commission = Buffer.from(item.commission, 'base64').toString('ascii');
      return item;
    }));
  }
});

const txRequest = new TransactionStatusRequest();
txRequest.setChainId(TX_CHAIN_ID);
txRequest.setTxHash(TX_HASH);
hubService.transactionStatus(txRequest, {}, function(err, response) {
  if (err) {
    console.log(err);
  } else {
    console.log(response.toObject().status);
  }
});

// same data but another syntax
hubService.transactionStatus(new TransactionStatusRequest([TX_CHAIN_ID, TX_HASH]), {}, function(err, response) {
  if (err) {
    console.log(err);
  } else {
    console.log(response.toObject().status);
  }
});

oracleService.prices(new QueryPricesRequest(), {}, function(err, response) {
  if (err) {
    console.log(err);
  } else {
    console.log(response.toObject().prices.listList);
  }
});

oracleService.ethFee(new QueryEthFeeRequest(), {}, function(err, response) {
  if (err) {
    console.log(err);
  } else {
    console.log(response.toObject());
  }
});

const {XMLHttpRequest} = require("xmlhttprequest");
global.XMLHttpRequest = XMLHttpRequest;

const {ParamsRequest, TokenInfosRequest} = require('../gen/mhub2/v1/query_pb.js')
const {QueryClient} = require('../gen/mhub2/v1/query_pb_service.js')


var queryService = new QueryClient('http://46.101.215.17:9091');


// queryService.params(request, {}, function(err, response) {
//   console.log({err,response}, response.toObject().params);
// });


queryService.tokenInfos(new TokenInfosRequest(), {}, function(err, response) {
  console.log({err,response});
});

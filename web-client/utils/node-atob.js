module.exports = function(str) {
    return Buffer.from(str, 'base64').toString('ascii');
}

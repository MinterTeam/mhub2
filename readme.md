# Minter Hub 2.0

## Build & Install

1. Install dependencies
```bash
apt-get update && \
  apt-get install -y git build-essential wget curl libssl-dev pkg-config
```

2. Install Golang
```bash
wget https://golang.org/dl/go1.17.3.linux-amd64.tar.gz && \
  rm -rf /usr/local/go && \
  tar -C /usr/local -xzf go1.17.3.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin:~/go/bin' >> ~/.profile
```

3. Install Rust
```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source ~/.profile
```

4. Clone Minter Hub repository
```bash
cd ~ && git clone https://github.com/MinterTeam/mhub2.git
```

5. Compile & install 
```bash
# Minter Hub node
cd ~/mhub2/module
make install

# Hub ↔ Minter oracle
cd ~/mhub2/minter-connector
make install

# Prices oracle
cd ~/mhub2/oracle
make install

# Keys generator
cd ~/mhub2/keys-generator
make install

# Hub ↔ Ethereum oracle
cd ~/mhub2/orchestrator
cargo install --locked --path orchestrator
```

## Run
1. Install and sync Minter Node 
```bash
minter node
```

2. Install and sync Ethereum & BSC node
```bash
geth --http --http.addr "127.0.0.1" --http.port "8545"
```

3. Sync Minter Hub Node
```bash
# Download genesis
mkdir -p ~/.mhub2/config/
curl ... > ~/.mhub2/config/genesis.json

# Start and sync Minter Hub node
mhub2 start \
  --p2p.persistent_peers="..."
```

for testnet:
```bash
# Download genesis
mkdir -p ~/.mhub/config/
curl https://raw.githubusercontent.com/MinterTeam/mhub2/master/testnet/genesis.json > ~/.mhub2/config/genesis.json

# Start and sync Minter Hub node
mhub2 start \
  --p2p.persistent_peers="9fb8440290a6180174e86e82bbba205af9d98779@46.101.215.17:26656"
```

4. Generate Hub account
```bash
mhub2 keys add validator1
```

- **WARNING: save generated key**
- Request some test HUB to your generated address

5. Create Hub validator
```bash
mhub2 tendermint show-validator # show validator's public key
mhub2 tx staking create-validator \
  --from=validator1 \
  --amount=1000000000000000000hub \
  --pubkey=<VALIDATOR PUBLIC KEY>  \
  --commission-max-change-rate="0.1" \
  --commission-max-rate="1" \
  --commission-rate="0.1" \
  --min-self-delegation="1" \
  --chain-id=mhub-mainnet-2 (mhub-testnet-11 for testnet)
```

- **WARNING: save tendermint validator's key**
- An important point: the validator is turned off if it does not commit data for a long time. You can turn in on again by sending an unjail transaction. Docs: `mhub2 tx slashing unjail --help`

6. Generate Minter & Ethereum keys
```bash
mhub-keys-generator
```
- **WARNING: save generated keys**
- Request some test ETH to your generated address

7. Register Ethereum keys
```bash
mhub2 query account <YOUR_ACCOUNT> # get account nonce

mhub-keys-generator make_delegate_sign <YOUR_ETH_PRIVATE_KEY> <YOUR_ACCOUNT> <YOUR_NONCE>

mhub2 tx mhub2 set-delegate-keys <VALADDR> <YOUR_ACCOUNT> <ETH_ADDR> <SIG> --from=...
```

8. Start services. *You can set them up as services or run in different terminal screens.*

- **Start Hub ↔ Ethereum oracle.** 
```
Ethereum Contract for mainnet: 0x1C03D454272382AAFF97b46Fe0f6cD66a65857C1
BSC Contract for mainnet: 0xdAc3632a0AFB486e8870018a86171109Dd110d2c

Ethereum Contract for testnet: 0xb150480c8f280C9A3f0CF1b7412dDaeF9eedb101
BSC Contract for testnet: 0xcD53640C87Acd89BD7935765167D1E6330201C89
```
```bash
orchestrator \
  --cosmos-phrase=<COSMOS MNEMONIC> \
  --ethereum-key=<ETHEREUM PRIVATE KEY> \
  --cosmos-grpc="http://127.0.0.1:9090" \
  --ethereum-rpc="http://127.0.0.1:8545/" \
  --fees=hub \
  --address-prefix=hub \
  --chain-id=ethereum \
  --contract-address=<ADDRESS OF ETHEREUM CONTRACT> 
```

```bash
orchestrator \
  --cosmos-phrase=<COSMOS MNEMONIC> \
  --ethereum-key=<ETHEREUM PRIVATE KEY> \
  --cosmos-grpc="http://127.0.0.1:9090" \
  --ethereum-rpc="http://127.0.0.1:8545/" \
  --fees=hub \
  --address-prefix=hub \
  --chain-id=bsc \
  --contract-address=<ADDRESS OF BSC CONTRACT> 
```

- **Start Hub ↔ Minter oracle.** 
```
Minter Multisig for testnet: Mxa7013f2c9a66db96a2712f0942784d4f9ba50378
Start Minter Block for testnet: 6647833

Minter Multisig for mainnet: Mx68f4839d7f32831b9234f9575f3b95e1afe21a56
Start Minter Block for mainnet: 7696500
```

```toml
# connector-config.toml

[minter]
# testnet|mainnet
chain = "mainnet"
multisig_addr = <ADDRESS OF MINTER MULTISIG>
private_key = <MINTER PRIVATE KEY>
api_addr = "http://127.0.0.1:8843/v2/"
start_block = <MINTER START BLOCK>
start_event_nonce = 1
start_batch_nonce = 1
start_valset_nonce = 0

[cosmos]
mnemonic = ""
grpc_addr = "127.0.0.1:9090"
rpc_addr = "http://127.0.0.1:26657"

```

```bash
mhub-minter-connector --config=connector-config.toml
```
  
- **Start price oracle**
```toml
# oracle-config.toml
holders_url = "https://explorer-hub-api.minter.network/api/tokens/1902/holders"
prices_url = "https://explorer-hub-api.minter.network/api/prices"

[cosmos]
mnemonic = <COSMOS MNEMONIC>
grpc_addr = "127.0.0.1:9090"
rpc_addr = "http://127.0.0.1:26657"
```

```bash
mhub-oracle --config=oracle-config.toml (--testnet)
``` 
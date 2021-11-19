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
cargo install --locked --path register_delegate_keys
```

## Run
1. Install and sync Minter Node 
```bash
minter node
```

2. Install and sync Ethereum node
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
curl https://raw.githubusercontent.com/MinterTeam/mhub2/dev/testnet-genesis.json > ~/.mhub/config/genesis.json

# Start and sync Minter Hub node
mhub2 start \
  --p2p.persistent_peers="..."
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
register-peggy-delegate-keys \
  --cosmos-phrase=<COSMOS MNEMONIC> \
  --validator-phrase=<COSMOS MNEMONIC> \
  --ethereum-key=<ETHEREUM PRIVATE KEY> \
  --address-prefix=hub \
  --cosmos-grpc="http://127.0.0.1:9090" \
  --fees=hub
```

8. Start services. *You can set them up as services or run in different terminal screens.*

- **Start Hub ↔ Ethereum oracle.** 
```
Ethereum Contract for testnet: ...
Ethereum Contract for mainnet: ...

BSC Contract for testnet: ...
BSC Contract for mainnet: ...
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
Minter Multisig for testnet: ...
Start Minter Block for testnet: ...

Minter Multisig for mainnet: ...
Start Minter Block for mainnet: ...
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

[minter]
api_addr = "http://127.0.0.1:8843/v2/"

[cosmos]
mnemonic = <COSMOS MNEMONIC>
grpc_addr = "127.0.0.1:9090"
rpc_addr = "http://127.0.0.1:26657"

[ethereum]
gas_price_providers = [
    "ethgasstation",
    "etherchain"
]
```

```bash
mhub-oracle --config=oracle-config.toml (--testnet)
``` 
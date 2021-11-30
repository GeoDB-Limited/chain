# odin-1 mainnet

## Genesis Time
The genesis transactions sent before {DATETIME} will be used to publish the genesis.json on or before {DATETIME} and then start the chain at 1400UTC. We will be announcing on all the platforms for the same. Please join our [Discord](https://discord.gg/cUXKyRq) and [Odin Telegram Group](https://t.me/odinprotocol) to stay updated.

### Hardware Requirements
#### Minimal
- 4 GB RAM
- 256 GB SSD
- 3.2 x4 GHz CPU

#### Recommended
- 8 GB RAM
- 512GB SSD
- 3.2 x4 GHz CPU

### Operating System
#### Recommended
- Linux(x86_64)

## Installation Steps


### Install Prerequisites 

#### Basic Packages
```bash:
# update the local package list and install any available upgrades 
sudo apt-get update && sudo apt upgrade -y 
# install toolchain and ensure accurate time synchronization 
sudo apt-get install make build-essential gcc git jq chrony -y
```

#### Install Go
Follow the instructions [here](https://golang.org/doc/install) to install Go.

For an Ubuntu LTS, you can do:
```bash:
wget https://golang.org/dl/go1.17.3.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.17.3.linux-amd64.tar.gz
```

Unless you want to configure in a non standard way, then set these in the `.profile` in the user's home (i.e. `~/`) folder.

```bash:
cat <<EOF >> ~/.profile
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export GO111MODULE=on
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
EOF
source ~/.profile
go version
```

### Install Odind from source

Download network
```bash:
git clone https://github.com/GeoDB-Limited/odin-core.git
cd odin-core
git fetch
git checkout <version-tag>
```

The `<version-tag>` will need to be set to either a testnet `chain-id` or the latest mainnet version tag.

For genesis, the mainnet version tag will be `v2.0.0` - i.e:
```bash:
git checkout v2.0.0
```

Once you're on the correct tag, you can build:

```bash:
# in odin-core dir
make install
```
	
To confirm that the installation has succeeded, you can run:

```bash:
odind version
# output: v2.0.0
```

### Init chain
```bash:
odind init $MONIKER_NAME --chain-id $CHAIN_ID
```

### Add/recover keys
```bash:
# To create new keypair - make sure you save the mnemonics!
odind keys add <key-name> 

# Restore existing odin wallet with mnemonic seed phrase. 
# You will be prompted to enter mnemonic seed. 
odind keys add <key-name> --recover 

# Add keys using ledger
odind keys show <key-name> --ledger

# Query the keystore for your public address 
odind keys show <key-name> -a
```

## Instructions for Genesis Validators

### Create Gentx

1. Add genesis account:
**WARNING: DO NOT PUT MORE THAN 10000000loki or your gentx will be rejected**

`odind add-genesis-account "{{KEY_NAME}}" 10000000loki`

2. Create Gentx
```
odind gentx "{{KEY_NAME}}" 10000000loki \
--chain-id odin-1 \
--moniker="{{VALIDATOR_NAME}}" \
--commission-max-change-rate=0.05 \
--commission-max-rate=0.20 \
--commission-rate=0.05 \
--details="XXXXXXXX" \
--security-contact="XXXXXXXX" \
--website="XXXXXXXX"
```

### Submit PR with Gentx and peer id

1. Copy the contents of ${HOME}/.odind/config/gentx/gentx-XXXXXXXX.json.

2. Fork the repository

3. Create a file gentx-{{VALIDATOR_NAME}}.json under the mainnets/odin-1/gentxs folder in the forked repo, paste the copied text into the file. Find reference file gentx-examplexxxxxxxx.json in the same folder.

4. Run `odind tendermint show-node-id` and copy your nodeID.

5. Run `ifconfig` or `curl ipinfo.io/ip` and copy your publicly reachable IP address.

6. Create a file peers-{{VALIDATOR_NAME}}.json under the mainnet/odin-1/peers folder in the forked repo, paste the copied text from the last two steps into the file. Find reference file sample-peers.txt in the same folder. (e.g. `fd4351c2e9928213b3d6ddce015c4664e6138@3.127.204.206:26656`)

7. Create a Pull Request to the main branch of the repository


## Instructions for non-Genesis Validators

### Add persistent peers
```bash:
PEERS = TBD
sed -i.bak -e "s/^persistent_peers *=.*/persistent_peers = \"$PEERS\"/" ~/.odind/config/config.toml
```

### Download genesis file
```bash:
curl TBD > ~/.odind/config/genesis.json
```

Verify the hash:
```
jq -S -c -M ' ' ~/.odind/config/genesis.json | shasum -a 256
```

### Sync the node
```bash:
odind start
```

### Create validator

```bash:
odind tx staking create-validator \ 
--amount 1000000loki \ 
--commission-max-change-rate "0.05" \ 
--commission-max-rate "0.10" \ 
--commission-rate "0.05" \ 
--min-self-delegation "1" \ 
--details "validators write bios too" \ 
--pubkey=$(odind tendermint show-validator) \ 
--moniker $MONIKER_NAME \ 
--chain-id $CHAIN_ID \ 
--gas-prices 0.025loki \ 
--from <key-name>
```

### Backup critical files
```bash:
priv_validator_key.json
node_key.json
```

### Setup Unit/Daemon file

```bash:
# 1. create daemon file
touch /etc/systemd/system/odin.service

# 2. run:
cat <<EOF >> /etc/systemd/system/odin.service
[UNIT]
Description=Odin daemon
After=network-online.target

[Service]
User=root
ExecStart=/root/go/bin/odind start
Restart=on-failure
RestartSec=3
LimitNOFILE=4096

[Install]
WantedBy=multi-user.target
EOF

# 3. reload the daemon
systemctl daemon-reload

# 4. enable service
systemctl enable odin.service

# 5. start daemon
systemctl start odin.service

# 6. watch logs
journalctl -u odin.service -f

```


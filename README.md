# `Chain`

Chain is based on the ibc-go `SimApp`, and always will be.  This will allow teams to test the latest and greatest of:

* Cosmos-SDK
* IBC-GO
* Comet-BFT

in an approchable and minimal format.  We're also satisfying a [the spec requirements for a wyoming dao](https://sos.wyo.gov/Forms/WyoBiz/DAO_Supplement.pdf).  

`Chain` is a CLI application built using the Cosmos SDK for testing and educational purposes.

## Wyoming DAO Compatibility

Chain implements a Wyoming DAO. 

The Cosmos SDK provides the foundational blockchain infrastructure that satisfies these requirements without requiring additional smart contract platforms like CosmWasm. Here's how Chain's architecture maps to the Wyoming DAO specifications:

### Blockchain Requirements (W.S. 34-29-106(g)(i))
Cosmos SDK provides a complete blockchain framework that meets Wyoming's definition of a blockchain as a "digital ledger of transactions." Chain inherits these properties:
- Immutable transaction records
- Cryptographically secured blocks
- Decentralized consensus mechanism via Comet BFT

### Smart Contract Requirements (W.S. 17-31-102(a)(ix))
While Chain doesn't use a separate smart contract platform, Cosmos SDK modules themselves function as "smart contracts" under Wyoming's definition:
- **Governance Module**: Functions as the automated transaction system for "administrating membership interest votes"
- **Bank Module**: Handles the "taking custody of and transferring" of digital assets
- **Staking Module**: Manages delegation relationships via code executed on the blockchain

Each of these modules executes deterministic code based on conditions specified in transactions, meeting the core definition of smart contracts.

### Management Structure (W.S. 17-31-109)
Cosmos SDK's governance module fully satisfies the requirement that "Management of a decentralized autonomous organization shall be vested in its members." Specifically:
- Token holders (members) submit and vote on proposals
- Proposals execute automatically when passed
- Governance parameters can be adjusted through the governance process itself

### Upgradability (W.S. 17-31-109)
The Cosmos SDK upgrade module ensures all chain code can be "updated, modified or otherwise upgraded" as required by Wyoming law:
- Governance proposals can include software upgrades
- Approved upgrades execute automatically at predetermined heights
- Ensures continuity while allowing evolution of the organization

### Membership Interests and Voting (W.S. 17-31-111)
The bank module tracks token ownership, which represents membership interests:
- Each token holder's voting power is proportional to their token balance
- The governance module enforces voting periods and tallying
- Delegation allows token holders to participate via trusted representatives

### Transparency (W.S. 17-31-112)
As an "open blockchain," Chain ensures all records are publicly accessible:
- Transaction history is transparent and immutable
- Token ownership is publicly verifiable
- Governance proposals and votes are publicly recorded

### Withdrawal of Members (W.S. 17-31-113)
Token holders can freely transfer their tokens, effectively implementing withdrawal:
- Members can transfer, sell, or alienate their tokens at any time
- This satisfies W.S. 17-31-113(d)(ii) for member withdrawal

### Dissolution Mechanisms (W.S. 17-31-114)
Governance proposals can implement any of the dissolution events specified in Wyoming law:
- A governance proposal can halt the chain
- If activity ceases for a year, it meets the legal condition for dissolution
- Chain governance can implement time-based or condition-based termination

To fully comply with Wyoming DAO requirements, Chain must be registered as a Wyoming LLC with articles of organization that include the required notices and statements. The actual blockchain serves as the technical implementation of the DAO's operations.

## Running Testnets with `chain`

Want to spin up a quick testnet with your friends? Follow these steps. Unless stated otherwise, all participants in the testnet must follow through with each step.

### 1. Download and Setup

Download IBC-go and unzip it. You can do this manually (via the GitHub UI) or with the git clone command:

```sh
git clone github.com/cosmos/ibc-go.git
```

Next, run this command to build the `chaind` binary in the `build` directory:

```sh
make build
```

Use the following command and skip all the next steps to configure your Chain node:

```sh
make init-chain
```

If you've run `chaind` in the past, you may need to reset your database before starting up a new testnet. You can do that with this command:

```sh
# you need to provide the moniker and chain ID
$ ./chaind init [moniker] --chain-id [chain-id]
```

The command should initialize a new working directory at the `~/.chain` location. 
The `moniker` and `chain-id` can be anything, but you must use the same `chain-id` subsequently.

### 2. Create a New Key

Execute this command to create a new key:

```sh
 ./chaind keys add [key_name]
```

⚠️ The command will create a new key with your chosen name.
Save the output somewhere safe; you'll need the address later.

### 3. Add Genesis Account

Add a genesis account to your testnet blockchain:

```sh
./chaind genesis add-genesis-account [key_name] [amount]
```

Where `key_name` is the same key name as before, and the `amount` is something like `10000000000000000000000000stake`.

### 4. Add the Genesis Transaction

This creates the genesis transaction for your testnet chain:

```sh
./chaind genesis gentx [key_name] [amount] --chain-id [chain-id]
```

The amount should be at least `1000000000stake`. Providing too much or too little may result in errors when you start your node.

### 5. Create the Genesis File

A participant must create the genesis file `genesis.json` with every participant's transaction. 
You can do this by gathering all the Genesis transactions under `config/gentx` and then executing this command:

```sh
./chaind genesis collect-gentxs
```

The command will create a new `genesis.json` file that includes data from all the validators. We sometimes call this the "super genesis file" to distinguish it from single-validator genesis files.

Once you've received the super genesis file, overwrite your original `genesis.json` file with the new super `genesis.json`.

Modify your `config/config.toml` (in the chain working directory) to include the other participants as persistent peers:

```toml
# Comma-separated list of nodes to keep persistent connections to
persistent_peers = "[validator_address]@[ip_address]:[port],[validator_address]@[ip_address]:[port]"
```

You can find `validator_address` by executing:

```sh
./chaind comet show-node-id
```

The output will be the hex-encoded `validator_address`. The default `port` is 26656.

### 6. Start the Nodes

Finally, execute this command to start your nodes:

```sh
./chaind start
```

Now you have a small testnet that you can use to try out changes to the Cosmos SDK or CometBFT!

> ⚠️ NOTE: Sometimes, creating the network through the `collect-gentxs` will fail, and validators will start in a funny state (and then panic).
> 
> If this happens, you can try to create and start the network first with a single validator and then add additional validators using a `create-validator` transaction.

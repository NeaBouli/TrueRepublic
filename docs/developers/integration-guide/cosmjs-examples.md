# CosmJS Integration Examples

Code examples for common blockchain operations using CosmJS.

## Setup

```bash
npm install @cosmjs/stargate @cosmjs/proto-signing @cosmjs/amino @cosmjs/encoding @cosmjs/crypto
```

## Connect to Chain (Read-Only)

```javascript
import { SigningStargateClient } from "@cosmjs/stargate";

const RPC_ENDPOINT = "https://rpc.truerepublic.network";

async function connect() {
    const client = await SigningStargateClient.connect(RPC_ENDPOINT);
    const height = await client.getHeight();
    console.log("Connected! Block height:", height);
    return client;
}
```

## Connect with Keplr (Signing)

```javascript
const CHAIN_ID = "truerepublic-1";

const chainConfig = {
    chainId: CHAIN_ID,
    chainName: "TrueRepublic",
    rpc: "https://rpc.truerepublic.network",
    rest: "https://lcd.truerepublic.network",
    bip44: { coinType: 118 },
    bech32Config: {
        bech32PrefixAccAddr: "truerepublic",
        bech32PrefixAccPub: "truerepublicpub",
        bech32PrefixValAddr: "truerepublicvaloper",
        bech32PrefixValPub: "truerepublicvaloperpub",
        bech32PrefixConsAddr: "truerepublicvalcons",
        bech32PrefixConsPub: "truerepublicvalconspub",
    },
    currencies: [
        { coinDenom: "PNYX", coinMinimalDenom: "pnyx", coinDecimals: 0 },
    ],
    feeCurrencies: [
        {
            coinDenom: "PNYX",
            coinMinimalDenom: "pnyx",
            coinDecimals: 0,
            gasPriceStep: { low: 0, average: 0, high: 0 },
        },
    ],
    stakeCurrency: {
        coinDenom: "PNYX",
        coinMinimalDenom: "pnyx",
        coinDecimals: 0,
    },
};

async function connectWithKeplr() {
    if (!window.keplr) throw new Error("Keplr not installed");

    // Suggest chain to Keplr (adds it if not present)
    await window.keplr.experimentalSuggestChain(chainConfig);
    await window.keplr.enable(CHAIN_ID);

    const offlineSigner = window.keplr.getOfflineSigner(CHAIN_ID);
    const accounts = await offlineSigner.getAccounts();
    const address = accounts[0].address;

    const client = await SigningStargateClient.connectWithSigner(
        "https://rpc.truerepublic.network",
        offlineSigner
    );

    return { client, address };
}
```

## Query Balance

```javascript
async function getBalance(address) {
    const client = await SigningStargateClient.connect(RPC_ENDPOINT);
    const balance = await client.getBalance(address, "pnyx");
    console.log(`Balance: ${balance.amount} ${balance.denom}`);
    return balance;
}
```

## Send Tokens

```javascript
async function sendPnyx(senderAddress, recipientAddress, amount) {
    const { client } = await connectWithKeplr();

    const result = await client.sendTokens(
        senderAddress,
        recipientAddress,
        [{ denom: "pnyx", amount: String(amount) }],
        "auto" // auto gas estimation
    );

    console.log("TX Hash:", result.transactionHash);
    return result;
}
```

## Query Domains (ABCI)

```javascript
async function fetchDomains() {
    const client = await SigningStargateClient.connect(RPC_ENDPOINT);
    const result = await client.queryAbci(
        "custom/truedemocracy/domains",
        new Uint8Array()
    );
    return JSON.parse(new TextDecoder().decode(result.value));
}

async function fetchDomain(name) {
    const client = await SigningStargateClient.connect(RPC_ENDPOINT);
    const result = await client.queryAbci(
        `custom/truedemocracy/domain/${name}`,
        new Uint8Array()
    );
    return JSON.parse(new TextDecoder().decode(result.value));
}
```

## Submit Governance Proposal

```javascript
async function submitProposal(domain, issue, suggestion, fee) {
    const { client, address } = await connectWithKeplr();

    const msg = {
        typeUrl: "/truedemocracy.MsgSubmitProposal",
        value: {
            sender: address,
            domain_name: domain,
            issue_name: issue,
            suggestion_name: suggestion,
            creator: address,
            fee: [{ denom: "pnyx", amount: String(fee) }],
            external_link: "",
        },
    };

    const result = await client.signAndBroadcast(address, [msg], "auto");
    console.log("Proposal TX:", result.transactionHash);
    return result;
}
```

## Place Stone (Vote)

```javascript
// Place stone on an issue
async function placeStoneOnIssue(domain, issue) {
    const { client, address } = await connectWithKeplr();

    const msg = {
        typeUrl: "/truedemocracy.MsgPlaceStoneOnIssue",
        value: {
            sender: address,
            domain_name: domain,
            issue_name: issue,
            member_addr: address,
        },
    };

    return client.signAndBroadcast(address, [msg], "auto");
}

// Place stone on a suggestion
async function placeStoneOnSuggestion(domain, issue, suggestion) {
    const { client, address } = await connectWithKeplr();

    const msg = {
        typeUrl: "/truedemocracy.MsgPlaceStoneOnSuggestion",
        value: {
            sender: address,
            domain_name: domain,
            issue_name: issue,
            suggestion_name: suggestion,
            member_addr: address,
        },
    };

    return client.signAndBroadcast(address, [msg], "auto");
}
```

## DEX Operations

```javascript
// Query pools
async function fetchPools() {
    const client = await SigningStargateClient.connect(RPC_ENDPOINT);
    const result = await client.queryAbci(
        "custom/dex/pools",
        new Uint8Array()
    );
    return JSON.parse(new TextDecoder().decode(result.value));
}

// Swap tokens
async function swap(inputDenom, inputAmount, outputDenom) {
    const { client, address } = await connectWithKeplr();

    const msg = {
        typeUrl: "/dex.MsgSwap",
        value: {
            sender: address,
            input_denom: inputDenom,
            input_amt: Number(inputAmount),
            output_denom: outputDenom,
        },
    };

    return client.signAndBroadcast(address, [msg], "auto");
}

// Add liquidity
async function addLiquidity(assetDenom, pnyxAmt, assetAmt) {
    const { client, address } = await connectWithKeplr();

    const msg = {
        typeUrl: "/dex.MsgAddLiquidity",
        value: {
            sender: address,
            asset_denom: assetDenom,
            pnyx_amt: Number(pnyxAmt),
            asset_amt: Number(assetAmt),
        },
    };

    return client.signAndBroadcast(address, [msg], "auto");
}
```

## Listen for Events (WebSocket)

```javascript
function subscribeToBlocks(callback) {
    const ws = new WebSocket("ws://localhost:26657/websocket");

    ws.onopen = () => {
        ws.send(JSON.stringify({
            jsonrpc: "2.0",
            method: "subscribe",
            id: 1,
            params: { query: "tm.event='NewBlock'" },
        }));
    };

    ws.onmessage = (event) => {
        const data = JSON.parse(event.data);
        if (data.result?.data?.value?.block) {
            callback(data.result.data.value.block);
        }
    };

    return () => ws.close();
}

// Usage
const unsubscribe = subscribeToBlocks((block) => {
    console.log("New block:", block.header.height);
});
```

## Error Handling

```javascript
async function safeBroadcast(client, address, msgs) {
    try {
        const result = await client.signAndBroadcast(address, msgs, "auto");

        if (result.code !== 0) {
            throw new Error(`Transaction failed: ${result.rawLog}`);
        }

        return result;
    } catch (error) {
        if (error.message.includes("insufficient funds")) {
            throw new Error("Not enough PNYX for this transaction");
        }
        if (error.message.includes("account sequence mismatch")) {
            throw new Error("Please wait for pending transaction to confirm");
        }
        throw error;
    }
}
```

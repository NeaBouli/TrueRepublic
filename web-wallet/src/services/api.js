import { SigningStargateClient } from "@cosmjs/stargate";

export const RPC_ENDPOINT = "https://rpc.truerepublic.network";
export const REST_ENDPOINT = "https://lcd.truerepublic.network";
export const CHAIN_ID = "truerepublic-1";

export const chainConfig = {
  chainId: CHAIN_ID,
  chainName: "TrueRepublic",
  rpc: RPC_ENDPOINT,
  rest: REST_ENDPOINT,
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
  stakeCurrency: { coinDenom: "PNYX", coinMinimalDenom: "pnyx", coinDecimals: 0 },
};

// Get a read-only client
export async function getQueryClient() {
  return SigningStargateClient.connect(RPC_ENDPOINT);
}

// Get a signing client (requires Keplr)
export async function getSigningClient() {
  if (!window.keplr) throw new Error("Keplr Wallet not installed!");
  const offlineSigner = window.keplr.getOfflineSigner(CHAIN_ID);
  return SigningStargateClient.connectWithSigner(RPC_ENDPOINT, offlineSigner);
}

// Query ABCI endpoint
export async function queryAbci(path) {
  const client = await getQueryClient();
  const result = await client.queryAbci(path, new Uint8Array());
  return JSON.parse(new TextDecoder().decode(result.value));
}

// Governance API
export async function fetchDomains() {
  return queryAbci("custom/truedemocracy/domains");
}

export async function submitProposal(sender, domainName, issueName, suggestionName) {
  const client = await getSigningClient();
  const msg = {
    typeUrl: "/truedemocracy.MsgSubmitProposal",
    value: {
      sender,
      domain_name: domainName,
      issue_name: issueName,
      suggestion_name: suggestionName,
      creator: sender,
      fee: [],
    },
  };
  return client.signAndBroadcast(sender, [msg], "auto");
}

export async function castVote(sender, domainName, issueName, suggestionName, stones) {
  const client = await getSigningClient();
  const msg = {
    typeUrl: "/truedemocracy.MsgCastVote",
    value: {
      sender,
      domain_name: domainName,
      issue_name: issueName,
      suggestion_name: suggestionName,
      stones: Number(stones),
    },
  };
  return client.signAndBroadcast(sender, [msg], "auto");
}

// Wallet API
export async function getBalance(address) {
  const client = await getQueryClient();
  return client.getBalance(address, "pnyx");
}

export async function sendTokens(sender, recipient, amount) {
  const client = await getSigningClient();
  return client.sendTokens(
    sender,
    recipient,
    [{ denom: "pnyx", amount: String(amount) }],
    "auto"
  );
}

// DEX API
export async function fetchPools() {
  return queryAbci("custom/dex/pools");
}

export async function swapTokens(sender, inputDenom, inputAmt, outputDenom) {
  const client = await getSigningClient();
  const msg = {
    typeUrl: "/dex.MsgSwap",
    value: {
      sender,
      input_denom: inputDenom,
      input_amt: Number(inputAmt),
      output_denom: outputDenom,
    },
  };
  return client.signAndBroadcast(sender, [msg], "auto");
}

// ZKP Voting API
export async function queryZKPState(domainName) {
  return queryAbci(`custom/truedemocracy/zkp_state/${domainName}`);
}

export async function queryNullifier(domainName, nullifierHash) {
  return queryAbci(`custom/truedemocracy/nullifier/${domainName}/${nullifierHash}`);
}

export async function queryPurgeSchedule(domainName) {
  return queryAbci(`custom/truedemocracy/purge_schedule/${domainName}`);
}

export async function submitAnonymousVote(sender, domainName, issueName, suggestionName, rating, proof, nullifierHash, merkleRoot) {
  const client = await getSigningClient();
  const msg = {
    typeUrl: "/truedemocracy.MsgRateWithProof",
    value: {
      sender,
      domain_name: domainName,
      issue_name: issueName,
      suggestion_name: suggestionName,
      rating: Number(rating),
      proof_hex: proof,
      nullifier_hash_hex: nullifierHash,
      merkle_root_hex: merkleRoot || "",
    },
  };
  return client.signAndBroadcast(sender, [msg], "auto");
}

// DEX Analytics API
export async function queryPoolStats(assetDenom) {
  return queryAbci(`custom/dex/pool_stats/${assetDenom}`);
}

export async function querySpotPrice(inputDenom, outputDenom) {
  return queryAbci(`custom/dex/spot_price/${inputDenom}/${outputDenom}`);
}

export async function queryLiquidityDepth(inputDenom, outputDenom) {
  return queryAbci(`custom/dex/liquidity_depth/${inputDenom}/${outputDenom}`);
}

export async function queryLPPosition(assetDenom, shares) {
  return queryAbci(`custom/dex/lp_position/${assetDenom}/${shares}`);
}

export async function queryEstimateSwap(inputDenom, amount, outputDenom) {
  return queryAbci(`custom/dex/estimate_swap/${inputDenom}/${amount}/${outputDenom}`);
}

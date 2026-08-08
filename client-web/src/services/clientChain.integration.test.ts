// @vitest-environment node
/// <reference types="node" />

import { generateKeyPairSync, sign } from 'node:crypto';
import { spawn, spawnSync, type ChildProcess } from 'node:child_process';
import { mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { createServer } from 'node:net';
import { get } from 'node:http';
import { tmpdir } from 'node:os';
import { resolve } from 'node:path';
import { DirectSecp256k1HdWallet } from '@cosmjs/proto-signing';
import { fromBech32 } from '@cosmjs/encoding';
import { StargateClient } from '@cosmjs/stargate';
import { afterAll, beforeAll, describe, expect, it } from 'vitest';
import { connectSigningClient, deliverMessages } from './signingClient';
import { createTxRegistry } from './txRegistry';
import { DEXService } from './dex';
import { GovernanceService } from './governance';
import { GovernanceTxService } from './governanceTx';
import { ZKPService } from './zkp';
import { ModuleQueryError } from './moduleQuery';
import type { ChainConfig } from '@/types/chain';

const enabled = process.env.TRUEREPUBLIC_CLIENT_CHAIN_INTEGRATION === '1';
const binary = resolve(
  process.env.TRUEREPUBLICD ?? '../build/truerepublicd'
);
const chainId = 'gh115-client-chain';
const gasPrice = '0.025upnyx';
const bootstrapOperator =
  'truerepublic13hgqwy9986x5nk6jt23ns5v7j0acs8qmhchhtw';

function freePort(): Promise<number> {
  return new Promise((resolvePort, reject) => {
    const server = createServer();
    server.once('error', reject);
    server.listen(0, '127.0.0.1', () => {
      const address = server.address();
      if (!address || typeof address === 'string') {
        server.close();
        reject(new Error('failed to allocate a local test port'));
        return;
      }
      const port = address.port;
      server.close((error) =>
        error ? reject(error) : resolvePort(port)
      );
    });
  });
}

function runBinary(args: string[], options?: { input?: string }): string {
  const result = spawnSync(binary, args, {
    encoding: 'utf8',
    input: options?.input,
    timeout: 60_000,
  });
  if (result.status !== 0) {
    throw new Error(
      `truerepublicd ${args.join(' ')} failed: ${result.stderr || result.stdout}`
    );
  }
  return result.stdout;
}

async function waitForRpc(rpcUrl: string): Promise<void> {
  const deadline = Date.now() + 60_000;
  let lastError = '';
  while (Date.now() < deadline) {
    try {
      const status = await new Promise<{ code: number; body: string }>(
        (resolveStatus, reject) => {
          const request = get(`${rpcUrl}/status`, (response) => {
            let body = '';
            response.setEncoding('utf8');
            response.on('data', (chunk) => {
              body += chunk;
            });
            response.on('end', () =>
              resolveStatus({ code: response.statusCode ?? 0, body })
            );
          });
          request.once('error', reject);
          request.setTimeout(2_000, () => {
            request.destroy(new Error('RPC readiness request timed out'));
          });
        }
      );
      if (status.code >= 200 && status.code < 300) {
        const payload = JSON.parse(status.body);
        const height = BigInt(
          payload.result?.sync_info?.latest_block_height ?? '0'
        );
        if (height >= 1n) return;
        lastError = 'RPC ready but first block is not committed';
      } else {
        lastError = `HTTP ${status.code}`;
      }
    } catch (error) {
      lastError = error instanceof Error ? error.message : String(error);
    }
    await new Promise((resolveWait) => setTimeout(resolveWait, 250));
  }
  throw new Error(`local chain RPC did not become ready: ${lastError}`);
}

async function waitForTx(rpcUrl: string, hash: string): Promise<void> {
  const client = await StargateClient.connect(rpcUrl);
  try {
    const deadline = Date.now() + 30_000;
    while (Date.now() < deadline) {
      const tx = await client.getTx(hash);
      if (tx) {
        expect(tx.code).toBe(0);
        return;
      }
      await new Promise((resolveWait) => setTimeout(resolveWait, 250));
    }
    throw new Error(`transaction ${hash} was not delivered`);
  } finally {
    client.disconnect();
  }
}

function rawEd25519PublicKey(publicKey: ReturnType<typeof generateKeyPairSync>['publicKey']): Buffer {
  const spki = publicKey.export({ format: 'der', type: 'spki' });
  return spki.subarray(spki.length - 32);
}

describe.skipIf(!enabled)('canonical client-to-chain delivery', () => {
  let home = '';
  let rpcUrl = '';
  let node: ChildProcess | undefined;
  let nodeLogs = '';
  let signerWallet: DirectSecp256k1HdWallet;
  let receiverAddress = '';

  async function cleanupResources(): Promise<void> {
    if (node && node.exitCode === null) {
      node.kill('SIGTERM');
      await new Promise<void>((resolveExit) => {
        const timer = setTimeout(() => {
          node?.kill('SIGKILL');
          resolveExit();
        }, 5_000);
        node?.once('exit', () => {
          clearTimeout(timer);
          resolveExit();
        });
      });
    }
    node = undefined;
    if (home) rmSync(home, { recursive: true, force: true });
    home = '';
  }

  beforeAll(async () => {
    try {
      home = mkdtempSync(`${tmpdir()}/truerepublic-gh115-`);
    signerWallet = await DirectSecp256k1HdWallet.generate(12, {
      prefix: 'truerepublic',
    });
    const receiver = await DirectSecp256k1HdWallet.generate(12, {
      prefix: 'truerepublic',
    });
    const [signerAccount] = await signerWallet.getAccounts();
    [receiverAddress] = (await receiver.getAccounts()).map(
      (account) => account.address
    );

    runBinary([
      'init',
      'gh115-client-chain',
      '--chain-id',
      chainId,
      '--home',
      home,
      '--bootstrap-operator',
      bootstrapOperator,
    ]);

    const genesisPath = resolve(home, 'config/genesis.json');
    const genesis = JSON.parse(readFileSync(genesisPath, 'utf8'));
    const auth = genesis.app_state.auth;
    const bank = genesis.app_state.bank;
    auth.accounts.push({
      '@type': '/cosmos.auth.v1beta1.BaseAccount',
      address: signerAccount.address,
      pub_key: null,
      account_number: '1',
      sequence: '0',
    });
    bank.balances.push({
      address: signerAccount.address,
      coins: [
        { denom: 'atom', amount: '2000000000000' },
        { denom: 'upnyx', amount: '2000000000000' },
      ],
    });
    const pnyxSupply = bank.supply.find(
      (coin: { denom: string }) => coin.denom === 'upnyx'
    );
    pnyxSupply.amount = (
      BigInt(pnyxSupply.amount) + 2_000_000_000_000n
    ).toString();
    bank.supply.unshift({ denom: 'atom', amount: '2000000000000' });
    writeFileSync(genesisPath, `${JSON.stringify(genesis, null, 2)}\n`, {
      mode: 0o600,
    });

    runBinary(
      [
        'keys',
        'add',
        'gh115-signer',
        '--recover',
        '--keyring-backend',
        'test',
        '--home',
        home,
        '--output',
        'json',
      ],
      { input: `${signerWallet.mnemonic}\n` }
    );

    // Allocate sequentially so the OS cannot immediately recycle one closed
    // probe socket into two different node listeners.
    const rpcPort = await freePort();
    const p2pPort = await freePort();
    const grpcPort = await freePort();
    const apiPort = await freePort();
    rpcUrl = `http://127.0.0.1:${rpcPort}`;
    node = spawn(binary, [
      'start',
      '--home',
      home,
      '--minimum-gas-prices',
      gasPrice,
      '--rpc.laddr',
      `tcp://127.0.0.1:${rpcPort}`,
      '--p2p.laddr',
      `tcp://127.0.0.1:${p2pPort}`,
      '--grpc.address',
      `127.0.0.1:${grpcPort}`,
      '--api.address',
      `tcp://127.0.0.1:${apiPort}`,
      '--log_level',
      'error',
    ]);
    node.stdout?.on('data', (chunk) => {
      nodeLogs = `${nodeLogs}${String(chunk)}`.slice(-20_000);
    });
    node.stderr?.on('data', (chunk) => {
      nodeLogs = `${nodeLogs}${String(chunk)}`.slice(-20_000);
    });
    await Promise.race([
      waitForRpc(rpcUrl),
      new Promise<never>((_, reject) => {
        node?.once('exit', (code, signal) =>
          reject(
            new Error(
              `local chain exited before RPC readiness (code=${code}, signal=${signal})\n${nodeLogs}`
            )
          )
        );
      }),
    ]);

    const createPoolOutput = runBinary([
      'tx',
      'dex',
      'create-pool',
      'atom',
      '100000000',
      '100000000',
      '--from',
      'gh115-signer',
      '--keyring-backend',
      'test',
      '--home',
      home,
      '--chain-id',
      chainId,
      '--node',
      `tcp://127.0.0.1:${rpcPort}`,
      '--gas',
      'auto',
      '--gas-adjustment',
      '1.3',
      '--gas-prices',
      gasPrice,
      '--yes',
      '--output',
      'json',
    ]);
    const createPool = JSON.parse(createPoolOutput);
      await waitForTx(rpcUrl, createPool.txhash);
    } catch (error) {
      await cleanupResources();
      throw error;
    }
  }, 120_000);

  afterAll(cleanupResources);

  it('simulates, signs, and delivers every maintained transaction family', async () => {
    const [account] = await signerWallet.getAccounts();
    const sender = fromBech32(account.address).data;
    const runtimeConfig: ChainConfig = {
      chainId,
      chainName: 'GH-121 local integration',
      rpc: rpcUrl,
      rest: '',
      bech32Prefix: 'truerepublic',
      coinDenom: 'PNYX',
      coinMinimalDenom: 'upnyx',
      coinDecimals: 6,
      gasPrice,
    };
    const client = await connectSigningClient(runtimeConfig, signerWallet);
    const deliver = async (typeUrl: string, value: object) => {
      try {
        return await deliverMessages(
          client,
          account.address,
          [{ typeUrl, value }],
          gasPrice
        );
      } catch (error) {
        throw new Error(
          `${typeUrl}: ${error instanceof Error ? error.message : String(error)}`
        );
      }
    };

    try {
      await expect(
        deliver('/cosmos.bank.v1beta1.MsgSend', {
          fromAddress: account.address,
          toAddress: receiverAddress,
          amount: [{ denom: 'upnyx', amount: '1' }],
        })
      ).resolves.toMatchObject({ success: true });

      await expect(
        deliver('/truedemocracy.MsgCreateDomain', {
          name: 'GH115',
          admin: sender,
          initialCoins: [{ denom: 'upnyx', amount: '1000000' }],
        })
      ).resolves.toMatchObject({ success: true });

      await expect(
        deliver('/truedemocracy.MsgAddMember', {
          sender,
          domainName: 'GH115',
          newMember: receiverAddress,
        })
      ).resolves.toMatchObject({ success: true });

      await expect(
        deliver('/truedemocracy.MsgSubmitProposal', {
          sender,
          domainName: 'GH115',
          issueName: 'Registry',
          suggestionName: 'Canonical',
          creator: account.address,
          fee: [{ denom: 'upnyx', amount: '5000' }],
          externalLink: '',
        })
      ).resolves.toMatchObject({ success: true });

      await expect(
        deliver('/truedemocracy.MsgPlaceStoneOnIssue', {
          sender,
          domainName: 'GH115',
          issueName: 'Registry',
          memberAddr: account.address,
        })
      ).resolves.toMatchObject({ success: true });

      await expect(
        deliver('/truedemocracy.MsgPlaceStoneOnSuggestion', {
          sender,
          domainName: 'GH115',
          issueName: 'Registry',
          suggestionName: 'Canonical',
          memberAddr: account.address,
        })
      ).resolves.toMatchObject({ success: true });

      const globalKey = generateKeyPairSync('ed25519');
      const domainKey = generateKeyPairSync('ed25519');
      const globalPubKeyHex = rawEd25519PublicKey(
        globalKey.publicKey
      ).toString('hex');
      const domainPubKeyHex = rawEd25519PublicKey(
        domainKey.publicKey
      ).toString('hex');
      const onboardingMessage = Buffer.from(
        `ONBOARD:${account.address}:GH115:${domainPubKeyHex}`
      );
      const signatureHex = sign(
        null,
        onboardingMessage,
        globalKey.privateKey
      ).toString('hex');

      await expect(
        deliver('/truedemocracy.MsgOnboardToDomain', {
          sender,
          domainName: 'GH115',
          domainPubKeyHex,
          globalPubKeyHex,
          signatureHex,
        })
      ).resolves.toMatchObject({ success: true });

      const identityCommitment = `${'0'.repeat(63)}1`;
      await expect(
        deliver('/truedemocracy.MsgRegisterIdentity', {
          sender,
          domainName: 'GH115',
          commitment: identityCommitment,
        })
      ).resolves.toMatchObject({ success: true });

      await expect(
        client.simulate(
          account.address,
          [
            {
              typeUrl: '/truedemocracy.MsgApproveOnboarding',
              value: {
                sender,
                domainName: 'GH115',
                requesterAddr: receiverAddress,
              },
            },
          ],
          ''
        )
      ).rejects.toThrow(/onboarding request not found/i);

      await expect(
        deliver('/dex.MsgAddLiquidity', {
          sender,
          assetDenom: 'atom',
          pnyxAmt: '10000000',
          assetAmt: '10000000',
        })
      ).resolves.toMatchObject({ success: true });

      await expect(
        deliver('/dex.MsgSwapExact', {
          sender,
          inputDenom: 'upnyx',
          inputAmt: '1000000',
          outputDenom: 'atom',
          minOutput: '1',
        })
      ).resolves.toMatchObject({ success: true });

      await expect(
        deliver('/dex.MsgRemoveLiquidity', {
          sender,
          assetDenom: 'atom',
          shares: '1',
        })
      ).resolves.toMatchObject({ success: true });

      const governance = new GovernanceService(runtimeConfig);
      const governanceTx = new GovernanceTxService(runtimeConfig);
      const dex = new DEXService(runtimeConfig);
      const zkp = new ZKPService(runtimeConfig);

      await expect(governance.listDomains()).resolves.toContainEqual(
        expect.objectContaining({ domainId: 'GH115', memberCount: 2 })
      );
      await expect(
        governance.listSuggestions('GH115', 'Registry')
      ).resolves.toContainEqual(
        expect.objectContaining({ suggestionId: 'Canonical' })
      );
      await expect(dex.listPools()).resolves.toContainEqual(
        expect.objectContaining({ asset_denom: 'atom' })
      );
      await expect(dex.getPoolStats('atom')).resolves.toMatchObject({
        asset_denom: 'atom',
      });
      await expect(
        dex.estimateSwap('upnyx', '1000', 'atom')
      ).resolves.toMatchObject({ hops: 1 });
      await expect(
        governanceTx.calculatePayToPut('GH115')
      ).resolves.toMatchObject({ domainMultiplier: 2 });
      const merkleProof = await zkp.fetchMerkleProof(
        'GH115',
        identityCommitment
      );
      expect(merkleProof).toMatchObject({
        root: expect.any(String),
        leaf: identityCommitment,
      });
      expect(merkleProof.pathIndices).toHaveLength(20);
      expect(merkleProof.pathElements).toHaveLength(20);
      await expect(dex.getPool('missing-denom')).rejects.toBeInstanceOf(
        ModuleQueryError
      );

      expect(() =>
        createTxRegistry().encode({ typeUrl: '/dex.MsgSwap', value: {} })
      ).toThrow(/Unregistered type url/);
    } catch (error) {
      throw new Error(
        `${error instanceof Error ? error.message : String(error)}\nLocal node logs:\n${nodeLogs}`
      );
    } finally {
      client.disconnect();
    }
  }, 120_000);
});

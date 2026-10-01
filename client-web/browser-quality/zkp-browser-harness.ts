import {
  fetchZKPArtifacts,
  parseZKPArtifactManifest,
} from '../src/services/zkpArtifacts';
import {
  ZKPWorkerProverRuntime,
  type ZKPWorkerPort,
} from '../src/services/zkpBrowserRuntime';
import {
  TestOnlyGroth16WasmProver,
  type TestOnlyZKPArtifacts,
} from '../src/services/zkpWasmProver';
import type { ProofInputs } from '../src/types/zkp';

interface GoldenVector {
  chain_id: string;
  domain_name: string;
  issue_name: string;
  suggestion_name: string;
  rating: number;
  synthetic_witness_hex: string;
  commitment_hex: string;
  merkle_root_hex: string;
  siblings_hex: string[];
  path_indices: number[];
  external_nullifier_hex: string;
}

interface CircuitSpec {
  vote_context_v2: {
    vector: { reward_recipient: string };
  };
}

interface BrowserHandoff {
  schema: 'truerepublic/zkp-wasm-handoff/v1';
  chainId: string;
  domainName: string;
  issueName: string;
  suggestionName: string;
  rating: number;
  rewardRecipient: string;
  proof: string;
  nullifierHash: string;
  merkleRoot: string;
  publicSignals: string[];
}

declare global {
  interface Window {
    runTrueRepublicZKPBrowserProof?: () => Promise<BrowserHandoff>;
  }
}

function workerFactory(wasm: Uint8Array): () => ZKPWorkerPort {
  return () => {
    // Module worker: the Go runtime glue is a fixed static import, never importScripts.
    const worker = new Worker('./zkp-browser-worker.js', { type: 'module' });
    const port: ZKPWorkerPort = {
      onmessage: null,
      onerror: null,
      postMessage: (message: unknown) => {
        if (typeof message !== 'object' || message === null) {
          throw new Error('worker request must be an object');
        }
        worker.postMessage({ ...message, wasm });
      },
      terminate: () => worker.terminate(),
    };
    worker.onmessage = (event: MessageEvent<unknown>) => {
      port.onmessage?.({ data: event.data });
    };
    worker.onerror = (event: ErrorEvent) => {
      port.onerror?.({ message: event.message });
    };
    return port;
  };
}

async function readJSON(path: string): Promise<unknown> {
  const response = await fetch(path, {
    credentials: 'omit',
    redirect: 'error',
    cache: 'no-store',
  });
  if (!response.ok) throw new Error(`fixture fetch failed: ${response.status}`);
  return response.json() as Promise<unknown>;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function requireString(record: Record<string, unknown>, key: string): string {
  const value = record[key];
  if (typeof value !== 'string' || value.length === 0) {
    throw new Error(`fixture field ${key} must be a non-empty string`);
  }
  return value;
}

function requireFieldHex(record: Record<string, unknown>, key: string): string {
  const value = requireString(record, key);
  if (!/^[0-9a-f]{64}$/u.test(value)) {
    throw new Error(`fixture field ${key} must be a canonical 32-byte hex value`);
  }
  return value;
}

function parseGoldenVector(value: unknown): GoldenVector {
  if (!isRecord(value)) throw new Error('golden vector must be an object');
  const siblings = value.siblings_hex;
  const indices = value.path_indices;
  const rating = value.rating;
  if (
    !Array.isArray(siblings) ||
    siblings.length !== 20 ||
    !siblings.every(
      (entry): entry is string =>
        typeof entry === 'string' && /^[0-9a-f]{64}$/u.test(entry)
    ) ||
    !Array.isArray(indices) ||
    indices.length !== 20 ||
    !indices.every((entry) => entry === 0 || entry === 1) ||
    !Number.isSafeInteger(rating) ||
    (rating as number) < -5 ||
    (rating as number) > 5
  ) {
    throw new Error('golden vector has an unsafe Merkle path or rating');
  }
  return {
    chain_id: requireString(value, 'chain_id'),
    domain_name: requireString(value, 'domain_name'),
    issue_name: requireString(value, 'issue_name'),
    suggestion_name: requireString(value, 'suggestion_name'),
    rating: rating as number,
    synthetic_witness_hex: requireFieldHex(value, 'synthetic_witness_hex'),
    commitment_hex: requireFieldHex(value, 'commitment_hex'),
    merkle_root_hex: requireFieldHex(value, 'merkle_root_hex'),
    siblings_hex: siblings,
    path_indices: indices as number[],
    external_nullifier_hex: requireFieldHex(value, 'external_nullifier_hex'),
  };
}

function parseCircuitSpec(value: unknown): CircuitSpec {
  if (!isRecord(value) || !isRecord(value.vote_context_v2)) {
    throw new Error('circuit specification has no v2 vote context');
  }
  const context = value.vote_context_v2;
  if (!isRecord(context.vector)) {
    throw new Error('circuit specification has no v2 vector');
  }
  return {
    vote_context_v2: {
      vector: {
        reward_recipient: requireString(context.vector, 'reward_recipient'),
      },
    },
  };
}

window.runTrueRepublicZKPBrowserProof = async () => {
  const manifest = parseZKPArtifactManifest(
    await readJSON('/__zkp/manifest.json')
  );
  const loaded = await fetchZKPArtifacts(manifest);
  const artifacts: TestOnlyZKPArtifacts = {
    constraintSystem: loaded.constraintSystem,
    provingKey: loaded.provingKey,
    verifyingKey: loaded.verifyingKey,
  };
  const runtime = new ZKPWorkerProverRuntime(
    workerFactory(loaded.wasm),
    artifacts
  );
  try {
    const vector = parseGoldenVector(
      await readJSON('/__zkp/golden_vector.json')
    );
    const spec = parseCircuitSpec(await readJSON('/__zkp/circuit.json'));
    const inputs: ProofInputs = {
      chainId: vector.chain_id,
      identitySecret: vector.synthetic_witness_hex,
      merkleRoot: vector.merkle_root_hex,
      merkleProof: {
        root: vector.merkle_root_hex,
        pathIndices: vector.path_indices,
        pathElements: vector.siblings_hex,
        leaf: vector.commitment_hex,
      },
      externalNullifier: vector.external_nullifier_hex,
      rating: vector.rating,
      domainName: vector.domain_name,
      issueName: vector.issue_name,
      suggestionName: vector.suggestion_name,
      rewardRecipient: spec.vote_context_v2.vector.reward_recipient,
    };
    const result = await new TestOnlyGroth16WasmProver(
      runtime,
      artifacts
    ).generate(inputs);
    return {
      schema: 'truerepublic/zkp-wasm-handoff/v1',
      chainId: inputs.chainId,
      domainName: inputs.domainName,
      issueName: inputs.issueName,
      suggestionName: inputs.suggestionName,
      rating: inputs.rating,
      rewardRecipient: inputs.rewardRecipient,
      proof: result.proof,
      nullifierHash: result.nullifierHash,
      merkleRoot: result.merkleRoot,
      publicSignals: result.publicSignals,
    };
  } finally {
    runtime.destroy();
    loaded.destroy();
  }
};

export {};

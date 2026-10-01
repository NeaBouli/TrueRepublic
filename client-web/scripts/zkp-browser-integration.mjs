import { createHash } from 'node:crypto';
import { chmodSync, readFileSync, writeFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { chromium } from 'playwright';
import { createServer as createViteServer } from 'vite';

const scriptDirectory = dirname(fileURLToPath(import.meta.url));
const clientDirectory = resolve(scriptDirectory, '..');
const repositoryDirectory = resolve(clientDirectory, '..');
const wasmPath = requiredEnvironment('TRUEREPUBLIC_ZKP_BROWSER_WASM_PATH');
const wasmExecPath = requiredEnvironment('TRUEREPUBLIC_WASM_EXEC_PATH');
const resultPath = requiredEnvironment('TRUEREPUBLIC_ZKP_RESULT_PATH');

// The Worker's Go runtime glue executes before its capability suppression, so
// only the exact trusted Go 1.26.6 lib/wasm/wasm_exec.js may be served.
const TRUSTED_WASM_EXEC_SHA256 = '0c949f4996f9a89698e4b5c586de32249c3b69b7baadb64d220073cc04acba14';
const wasmExecDigest = createHash('sha256').update(readFileSync(wasmExecPath)).digest('hex');
if (wasmExecDigest !== TRUSTED_WASM_EXEC_SHA256) {
  throw new Error(
    `untrusted Go wasm_exec.js glue: sha256 ${wasmExecDigest} is not the pinned Go 1.26.6 digest`
  );
}

const routes = new Map([
  ['/__zkp/manifest.json', resolve(repositoryDirectory, 'configs/security/zkp-browser-artifacts.json')],
  ['/__zkp/membership_v2.cs', resolve(repositoryDirectory, 'x/truedemocracy/testdata/zkp/membership_v2.cs')],
  ['/__zkp/membership_v2.pk', resolve(repositoryDirectory, 'x/truedemocracy/testdata/zkp/membership_v2.pk')],
  ['/__zkp/membership_v2.vk', resolve(repositoryDirectory, 'x/truedemocracy/testdata/zkp/membership_v2.vk')],
  ['/__zkp/golden_vector.json', resolve(repositoryDirectory, 'x/truedemocracy/testdata/zkp/golden_vector.json')],
  ['/__zkp/circuit.json', resolve(repositoryDirectory, 'configs/security/zkp-circuit.json')],
  ['/__zkp/zkp-prover.wasm', wasmPath],
  ['/__zkp/wasm_exec.js', wasmExecPath],
]);

const artifactPlugin = {
  name: 'truerepublic-test-only-zkp-browser-artifacts',
  // The module worker statically imports /__zkp/wasm_exec.js; keep that URL
  // external so Vite serves it from this middleware instead of resolving a file.
  resolveId(id) {
    return routes.has(id) ? { id, external: true } : null;
  },
  configureServer(server) {
    server.middlewares.use((request, response, next) => {
      const path = new URL(request.url ?? '/', 'http://127.0.0.1').pathname;
      const file = routes.get(path);
      if (file === undefined) {
        next();
        return;
      }
      const data = readFileSync(file);
      response.statusCode = 200;
      response.setHeader('Content-Length', String(data.byteLength));
      response.setHeader(
        'Content-Type',
        path.endsWith('.json')
          ? 'application/json'
          : path.endsWith('.js')
            ? 'application/javascript'
            : path.endsWith('.wasm')
              ? 'application/wasm'
              : 'application/octet-stream'
      );
      response.end(data);
    });
  },
};

const vite = await createViteServer({
  root: clientDirectory,
  configFile: resolve(clientDirectory, 'vite.config.ts'),
  plugins: [artifactPlugin],
  server: { host: '127.0.0.1', port: 0, strictPort: false },
});

let browser;
try {
  await vite.listen();
  const address = vite.httpServer?.address();
  if (address === null || typeof address !== 'object') {
    throw new Error('Vite did not expose a loopback port');
  }
  browser = await chromium.launch({ headless: true });
  const page = await browser.newPage();
  page.setDefaultTimeout(120_000);
  const browserErrors = [];
  page.on('pageerror', (error) => browserErrors.push(error.message));
  page.on('console', (message) => {
    if (message.type() === 'error') browserErrors.push(message.text());
  });
  await page.goto(
    `http://127.0.0.1:${address.port}/browser-quality/zkp-browser-harness.html`,
    { waitUntil: 'networkidle' }
  );
  await page.waitForFunction(
    () => typeof window.runTrueRepublicZKPBrowserProof === 'function'
  );
  const handoff = await Promise.race([
    page.evaluate(() => window.runTrueRepublicZKPBrowserProof?.()),
    new Promise((_, reject) =>
      setTimeout(
        () => reject(new Error('browser proof generation timed out')),
        120_000
      )
    ),
  ]);
  if (
    handoff === undefined ||
    handoff.schema !== 'truerepublic/zkp-wasm-handoff/v1' ||
    typeof handoff.proof !== 'string' ||
    handoff.proof.length !== 328 ||
    !Array.isArray(handoff.publicSignals) ||
    handoff.publicSignals.length !== 4 ||
    !handoff.publicSignals.every(
      (signal) => typeof signal === 'string' && /^[0-9a-f]{64}$/u.test(signal)
    ) ||
    browserErrors.length !== 0
  ) {
    throw new Error(
      browserErrors.length === 0
        ? 'browser prover returned a malformed keeper handoff'
        : `browser emitted errors: ${browserErrors.join(' | ')}`
    );
  }
  writeFileSync(resultPath, `${JSON.stringify(handoff)}\n`, {
    encoding: 'utf8',
    mode: 0o600,
  });
  chmodSync(resultPath, 0o600);
  process.stdout.write('Real Chromium Go/WASM proof generation passed.\n');
} finally {
  await browser?.close();
  await vite.close();
}

function requiredEnvironment(name) {
  const value = process.env[name];
  if (!value) throw new Error(`${name} is required`);
  return value;
}

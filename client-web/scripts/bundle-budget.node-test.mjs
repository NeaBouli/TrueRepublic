import test from 'node:test';
import assert from 'node:assert/strict';
import { evaluateBundleBudget } from './bundle-budget.mjs';

const budget = {
  routeEntries: 2,
  entry: { raw: 100, gzip: 50 },
  route: { raw: 40, gzip: 20 },
  chunk: { raw: 120, gzip: 60 },
  total: { raw: 200, gzip: 100 },
};

const manifest = {
  'index.html': { file: 'assets/index.js', isEntry: true },
  'src/components/one.tsx': { file: 'assets/one.js', isDynamicEntry: true },
  'src/components/two.tsx': { file: 'assets/two.js', isDynamicEntry: true },
};

const files = {
  'assets/index.js': { raw: 80, gzip: 40 },
  'assets/one.js': { raw: 30, gzip: 15 },
  'assets/two.js': { raw: 20, gzip: 10 },
};

test('accepts a complete lazy-route build within every budget', () => {
  const result = evaluateBundleBudget(manifest, files, budget);

  assert.equal(result.ok, true);
  assert.deepEqual(result.violations, []);
  assert.deepEqual(result.measurements, {
    entry: { raw: 80, gzip: 40 },
    routeEntries: 2,
    maxRoute: 15,
    total: { raw: 130, gzip: 65 },
  });
});

test('fails closed when the manifest loses an entry or lazy route', () => {
  const incompleteManifest = {
    'src/components/one.tsx': manifest['src/components/one.tsx'],
  };
  const result = evaluateBundleBudget(incompleteManifest, files, budget);

  assert.equal(result.ok, false);
  assert.ok(result.violations.some((violation) => violation.includes('application entry')));
  assert.ok(result.violations.some((violation) => violation.includes('lazy route entries')));
});

test('reports entry, route, chunk, and total regressions independently', () => {
  const oversizedFiles = {
    ...files,
    'assets/index.js': { raw: 121, gzip: 61 },
    'assets/one.js': { raw: 60, gzip: 30 },
  };
  const result = evaluateBundleBudget(manifest, oversizedFiles, budget);

  assert.equal(result.ok, false);
  assert.ok(result.violations.some((violation) => violation.startsWith('entry:')));
  assert.ok(result.violations.some((violation) => violation.startsWith('route ')));
  assert.ok(result.violations.some((violation) => violation.startsWith('chunk ')));
  assert.ok(result.violations.some((violation) => violation.startsWith('total JavaScript:')));
});

test('fails closed when a manifest artifact is missing', () => {
  const missing = { ...files };
  delete missing['assets/two.js'];
  const result = evaluateBundleBudget(manifest, missing, budget);

  assert.equal(result.ok, false);
  assert.ok(result.violations.some((violation) => violation.includes('emitted file is missing')));
});

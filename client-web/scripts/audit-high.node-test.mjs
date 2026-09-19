import assert from 'node:assert/strict';
import test from 'node:test';
import { spawnSync } from 'node:child_process';
import {
  classifyAuditResult,
  evaluateAudit,
  evaluateRouterBoundary,
  isPlainObject,
  safeDiagnostic,
} from './audit-high.mjs';

const cleanReport = JSON.stringify({ vulnerabilities: {}, metadata: {} });
const advisoryReport = JSON.stringify({
  vulnerabilities: {
    'react-router': {
      severity: 'high',
      via: [
        {
          name: 'react-router',
          severity: 'high',
          url: 'https://github.com/advisories/GHSA-qwww-vcr4-c8h2',
        },
      ],
    },
  },
});

function spawnResult(overrides) {
  return { status: 0, signal: null, stdout: '', stderr: '', error: undefined, ...overrides };
}

const allowedAdvisory = {
  name: 'react-router',
  severity: 'high',
  url: 'https://github.com/advisories/GHSA-qwww-vcr4-c8h2',
};

test('passes an audit report without blocking advisories', () => {
  assert.deepEqual(evaluateAudit({ vulnerabilities: {} }), {
    ok: true,
    accepted: [],
    blockers: [],
  });
});

test('accepts only the exact RSC advisory and its dependent package', () => {
  const result = evaluateAudit({
    vulnerabilities: {
      'react-router': { severity: 'high', via: [allowedAdvisory] },
      'react-router-dom': { severity: 'high', via: ['react-router'] },
    },
  });

  assert.equal(result.ok, true);
  assert.deepEqual(result.accepted.sort(), ['react-router', 'react-router-dom']);
  assert.deepEqual(result.blockers, []);
});

test('fails closed for any additional high advisory', () => {
  const result = evaluateAudit({
    vulnerabilities: {
      'react-router': { severity: 'high', via: [allowedAdvisory] },
      malicious: {
        severity: 'critical',
        via: [{ severity: 'critical', url: 'https://example.test/not-allowed' }],
      },
    },
  });

  assert.equal(result.ok, false);
  assert.deepEqual(result.blockers, ['malicious']);
});

test('fails closed for malformed and cyclic dependency chains', () => {
  assert.equal(evaluateAudit(null).ok, false);
  assert.equal(
    evaluateAudit({
      vulnerabilities: {
        first: { severity: 'high', via: ['second'] },
        second: { severity: 'high', via: ['first'] },
      },
    }).ok,
    false
  );
});

test('classifies a valid clean report as evaluable', () => {
  const outcome = classifyAuditResult(spawnResult({ stdout: cleanReport }));

  assert.equal(outcome.kind, 'report');
  assert.deepEqual(evaluateAudit(outcome.report), { ok: true, accepted: [], blockers: [] });
});

test('evaluates a valid advisory report even when npm exits nonzero', () => {
  const outcome = classifyAuditResult(spawnResult({ status: 1, stdout: advisoryReport }));

  assert.equal(outcome.kind, 'report');
  assert.deepEqual(evaluateAudit(outcome.report), {
    ok: true,
    accepted: ['react-router'],
    blockers: [],
  });
});

test('fails closed for contradictory or invalid process status', () => {
  assert.equal(
    classifyAuditResult(spawnResult({ status: 1, stdout: cleanReport })).kind,
    'operational'
  );
  assert.equal(
    classifyAuditResult(spawnResult({ status: 2, stdout: advisoryReport })).kind,
    'operational'
  );
  assert.equal(
    classifyAuditResult(spawnResult({ status: null, stdout: cleanReport })).kind,
    'spawn-failure'
  );
});

test('classifies npm operational error JSON separately from a report', () => {
  const outcome = classifyAuditResult(
    spawnResult({
      status: 1,
      stdout: JSON.stringify({
        error: {
          code: 'ENOTFOUND',
          summary: 'request to https://user:hunter2@registry.test/ failed',
          detail: 'network error',
        },
      }),
    })
  );

  assert.equal(outcome.kind, 'operational');
  assert.match(outcome.detail, /ENOTFOUND/);
  assert.doesNotMatch(outcome.detail, /hunter2/);
});

test('never echoes unbounded or credential-bearing npm diagnostics', () => {
  assert.equal(safeDiagnostic('E401'), 'E401');
  assert.equal(safeDiagnostic('https://user:hunter2@registry.test/'), 'redacted');
  assert.equal(safeDiagnostic('x'.repeat(5000)), 'redacted');
  assert.equal(safeDiagnostic(undefined), 'unavailable');
  assert.equal(safeDiagnostic('   '), 'unavailable');
});

test('classifies malformed npm output as malformed, not as a report', () => {
  const cases = [
    spawnResult({ stdout: 'npm warn config ...\nnot json at all' }),
    spawnResult({ stdout: '{"vulnerabilities":' }),
    spawnResult({ stdout: '' }),
    spawnResult({ stdout: '   ' }),
    spawnResult({ stdout: 'null' }),
    spawnResult({ stdout: '[]' }),
    spawnResult({ stdout: JSON.stringify({ metadata: {} }) }),
    spawnResult({ stdout: JSON.stringify({ vulnerabilities: [] }) }),
    spawnResult({ stdout: JSON.stringify({ vulnerabilities: 'none' }) }),
    spawnResult({ stdout: JSON.stringify({ vulnerabilities: null }) }),
  ];

  for (const result of cases) {
    assert.equal(classifyAuditResult(result).kind, 'malformed', result.stdout);
  }
});

test('classifies spawn and process failures as spawn failures', () => {
  const missingBinary = spawnSync('truerepublic-npm-that-does-not-exist', ['audit', '--json'], {
    encoding: 'utf8',
  });
  const spawnFailure = classifyAuditResult(missingBinary);
  assert.equal(spawnFailure.kind, 'spawn-failure');
  assert.match(spawnFailure.detail, /ENOENT/);

  assert.equal(
    classifyAuditResult(spawnResult({ signal: 'SIGKILL', stdout: cleanReport })).kind,
    'spawn-failure'
  );
  assert.equal(
    classifyAuditResult(
      spawnResult({ error: Object.assign(new Error('boom'), { code: 'ENOBUFS' }), stdout: '' })
    ).kind,
    'spawn-failure'
  );
  assert.equal(classifyAuditResult(undefined).kind, 'spawn-failure');
  assert.equal(classifyAuditResult(null).kind, 'spawn-failure');
});

test('a spawn or operational failure is never evaluated as a clean audit', () => {
  const failures = [
    spawnResult({ error: Object.assign(new Error('boom'), { code: 'ENOENT' }) }),
    spawnResult({ stdout: JSON.stringify({ error: { code: 'E401' } }) }),
    spawnResult({ stdout: 'not json' }),
  ];

  for (const result of failures) {
    const outcome = classifyAuditResult(result);
    assert.notEqual(outcome.kind, 'report');
    assert.equal(outcome.report, undefined);
    assert.equal(evaluateAudit(outcome.report).ok, false);
  }
});

test('rejects non-plain and prototype-borrowed vulnerability shapes', () => {
  assert.equal(isPlainObject(Object.create(null)), true);
  assert.equal(isPlainObject([]), false);
  assert.equal(isPlainObject(new Error('x')), false);

  assert.equal(evaluateAudit({ vulnerabilities: [] }).ok, false);
  assert.deepEqual(evaluateAudit({ vulnerabilities: { pkg: 'high' } }), {
    ok: false,
    accepted: [],
    blockers: ['pkg'],
  });
  assert.equal(evaluateAudit({ vulnerabilities: { pkg: { severity: 7 } } }).ok, false);
  assert.equal(evaluateAudit({ vulnerabilities: { pkg: { severity: 'low' } } }).ok, true);
  assert.equal(
    evaluateAudit({
      vulnerabilities: { pkg: { severity: 'high', via: ['constructor'] } },
    }).ok,
    false
  );
});

test('permits only the reviewed declarative SPA router surface', () => {
  assert.deepEqual(
    evaluateRouterBoundary([
      {
        path: 'App.tsx',
        source: "import { BrowserRouter, Routes } from 'react-router-dom';",
      },
      {
        path: 'routes.tsx',
        source: "import type { RouteObject } from 'react-router-dom';",
      },
    ]),
    []
  );
});

test('rejects data-router, RSC-capable, and alternate router imports', () => {
  assert.deepEqual(
    evaluateRouterBoundary([
      {
        path: 'data-router.tsx',
        source: "import { RouterProvider } from 'react-router-dom';",
      },
      {
        path: 'server.ts',
        source: "import { createStaticHandler } from 'react-router';",
      },
      {
        path: 'namespace.ts',
        source: "import * as Router from 'react-router-dom';",
      },
      {
        path: 'dynamic.ts',
        source: "const Router = await import('react-router-dom');",
      },
      {
        path: 'barrel.ts',
        source: "export { RouterProvider } from 'react-router-dom';",
      },
      {
        path: 'template.ts',
        source: 'const Router = await import(`react-router-dom`);',
      },
      {
        path: 'computed.ts',
        source: "const moduleName = 'react-router-dom'; const Router = await import(moduleName);",
      },
    ]),
    [
      'data-router.tsx: disallowed router API RouterProvider',
      'server.ts: disallowed router module react-router',
      'namespace.ts: router namespace imports are disallowed',
      'dynamic.ts: router dynamic imports must use reviewed named APIs',
      'barrel.ts: disallowed router API RouterProvider',
      'template.ts: router dynamic imports must use reviewed named APIs',
      'computed.ts: computed module imports are disallowed',
    ]
  );
});

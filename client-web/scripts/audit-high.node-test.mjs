import assert from 'node:assert/strict';
import test from 'node:test';
import { evaluateAudit, evaluateRouterBoundary } from './audit-high.mjs';

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

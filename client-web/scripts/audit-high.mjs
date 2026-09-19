import { spawnSync } from 'node:child_process';
import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';
import ts from 'typescript';

export const ALLOWED_ADVISORIES = new Set([
  // React Router's RSC/server-action path is not present in this BrowserRouter
  // SPA. Remove this exception when React 19 + react-router >=8.3.0 lands.
  'https://github.com/advisories/GHSA-qwww-vcr4-c8h2',
]);

const BLOCKING_SEVERITIES = new Set(['high', 'critical']);

// npm audit JSON for a large dependency tree exceeds the 1 MiB spawnSync
// default. The cap stays bounded so a runaway child cannot exhaust memory;
// exceeding it surfaces as a spawn failure and therefore fails closed.
const AUDIT_MAX_BUFFER_BYTES = 32 * 1024 * 1024;
const MAX_DIAGNOSTIC_CHARS = 400;
const SAFE_CODE_PATTERN = /^[A-Za-z0-9_.-]{1,64}$/;
const ROUTER_IMPORT_ALLOWLIST = new Set([
  'BrowserRouter',
  'MemoryRouter',
  'Navigate',
  'Route',
  'RouteObject',
  'Routes',
  'matchRoutes',
  'useLocation',
  'useNavigate',
  'useParams',
  'useSearchParams',
]);

function advisoryIsAllowed(via) {
  return (
    typeof via === 'object' &&
    via !== null &&
    typeof via.url === 'string' &&
    via.name === 'react-router' &&
    ALLOWED_ADVISORIES.has(via.url)
  );
}

export function evaluateRouterBoundary(files) {
  const violations = [];

  for (const { path, source } of files) {
    const scriptKind = path.endsWith('x') ? ts.ScriptKind.TSX : ts.ScriptKind.TS;
    const sourceFile = ts.createSourceFile(path, source, ts.ScriptTarget.Latest, true, scriptKind);

    const checkNames = (names) => {
      for (const name of names) {
        if (!ROUTER_IMPORT_ALLOWLIST.has(name)) {
          violations.push(`${path}: disallowed router API ${name}`);
        }
      }
    };

    const checkModule = (moduleName) => {
      if (!moduleName.startsWith('react-router')) return false;
      if (moduleName !== 'react-router-dom') {
        violations.push(`${path}: disallowed router module ${moduleName}`);
        return false;
      }
      return true;
    };

    const visit = (node) => {
      if (ts.isImportDeclaration(node) && ts.isStringLiteral(node.moduleSpecifier)) {
        const moduleName = node.moduleSpecifier.text;
        if (checkModule(moduleName)) {
          const clause = node.importClause;
          if (!clause || clause.name || !clause.namedBindings) {
            violations.push(`${path}: router imports must use reviewed named APIs`);
          } else if (ts.isNamespaceImport(clause.namedBindings)) {
            violations.push(`${path}: router namespace imports are disallowed`);
          } else {
            checkNames(
              clause.namedBindings.elements.map((element) =>
                (element.propertyName ?? element.name).text
              )
            );
          }
        }
      }

      if (ts.isExportDeclaration(node) && node.moduleSpecifier) {
        const moduleName = ts.isStringLiteral(node.moduleSpecifier)
          ? node.moduleSpecifier.text
          : '';
        if (checkModule(moduleName)) {
          if (!node.exportClause || !ts.isNamedExports(node.exportClause)) {
            violations.push(`${path}: router wildcard/namespace exports are disallowed`);
          } else {
            checkNames(
              node.exportClause.elements.map((element) =>
                (element.propertyName ?? element.name).text
              )
            );
          }
        }
      }

      if (
        ts.isCallExpression(node) &&
        (node.expression.kind === ts.SyntaxKind.ImportKeyword ||
          (ts.isIdentifier(node.expression) && node.expression.text === 'require'))
      ) {
        if (
          node.arguments.length !== 1 ||
          (!ts.isStringLiteral(node.arguments[0]) &&
            !ts.isNoSubstitutionTemplateLiteral(node.arguments[0]))
        ) {
          violations.push(`${path}: computed module imports are disallowed`);
        } else {
          const moduleName = node.arguments[0].text;
          if (checkModule(moduleName)) {
            const declaration = node.parent?.parent;
            if (
              node.expression.kind !== ts.SyntaxKind.ImportKeyword ||
              !ts.isAwaitExpression(node.parent) ||
              !declaration ||
              !ts.isVariableDeclaration(declaration) ||
              !ts.isObjectBindingPattern(declaration.name)
            ) {
              violations.push(`${path}: router dynamic imports must use reviewed named APIs`);
            } else {
              checkNames(
                declaration.name.elements.map((element) =>
                  (element.propertyName ?? element.name).getText(sourceFile)
                )
              );
            }
          }
        }
      }

      ts.forEachChild(node, visit);
    };

    visit(sourceFile);
  }

  return violations;
}

function sourceFiles(directory) {
  const files = [];
  for (const entry of readdirSync(directory, { withFileTypes: true })) {
    const path = join(directory, entry.name);
    if (entry.isDirectory()) files.push(...sourceFiles(path));
    if (entry.isFile() && /\.[cm]?[jt]sx?$/.test(entry.name)) {
      files.push({ path, source: readFileSync(path, 'utf8') });
    }
  }
  return files;
}

export function isPlainObject(value) {
  if (typeof value !== 'object' || value === null || Array.isArray(value)) return false;
  const prototype = Object.getPrototypeOf(value);
  return prototype === Object.prototype || prototype === null;
}

// npm error output can embed registry URLs carrying credentials. Only short,
// well-formed machine codes are echoed; everything else is dropped so the gate
// never turns a failure into a secret disclosure.
export function safeDiagnostic(value) {
  if (typeof value !== 'string') return 'unavailable';
  const trimmed = value.trim();
  if (trimmed === '') return 'unavailable';
  if (!SAFE_CODE_PATTERN.test(trimmed)) return 'redacted';
  return trimmed.slice(0, MAX_DIAGNOSTIC_CHARS);
}

function auditReportIsOperationalError(parsed) {
  return Object.hasOwn(parsed, 'error');
}

/**
 * Classify a spawnSync result into exactly one outcome. Only `report` may be
 * evaluated; every other outcome fails the gate closed. A nonzero npm exit
 * status is deliberately not an outcome of its own: `npm audit` exits nonzero
 * whenever it finds advisories at or above the audit level, and those runs
 * still carry a valid report that must be evaluated normally.
 */
export function classifyAuditResult(result) {
  if (typeof result !== 'object' || result === null) {
    return { kind: 'spawn-failure', detail: 'npm audit returned no result' };
  }

  if (result.error) {
    return {
      kind: 'spawn-failure',
      detail: `npm audit could not be executed (${safeDiagnostic(result.error.code)})`,
    };
  }

  if (result.signal) {
    return {
      kind: 'spawn-failure',
      detail: `npm audit was terminated by signal ${safeDiagnostic(result.signal)}`,
    };
  }

  if (typeof result.stdout !== 'string' || result.stdout.trim() === '') {
    return { kind: 'malformed', detail: 'npm audit produced no JSON on stdout' };
  }

  let parsed;
  try {
    parsed = JSON.parse(result.stdout);
  } catch {
    return { kind: 'malformed', detail: 'npm audit output was not valid JSON' };
  }

  if (!isPlainObject(parsed)) {
    return { kind: 'malformed', detail: 'npm audit output was not a JSON object' };
  }

  if (auditReportIsOperationalError(parsed)) {
    const code = isPlainObject(parsed.error)
      ? safeDiagnostic(parsed.error.code)
      : safeDiagnostic(parsed.error);
    return { kind: 'operational', detail: `npm audit reported an operational error (${code})` };
  }

  if (!isPlainObject(parsed.vulnerabilities)) {
    return {
      kind: 'malformed',
      detail: 'npm audit report has no valid vulnerabilities object',
    };
  }

  if (!Number.isInteger(result.status) || result.status < 0) {
    return { kind: 'spawn-failure', detail: 'npm audit returned no valid exit status' };
  }

  if (result.status > 1) {
    return {
      kind: 'operational',
      detail: `npm audit exited with unexpected status ${result.status}`,
    };
  }

  if (result.status === 1 && Object.keys(parsed.vulnerabilities).length === 0) {
    return {
      kind: 'operational',
      detail: 'npm audit exited unsuccessfully without reporting an advisory',
    };
  }

  return { kind: 'report', report: parsed };
}

function vulnerabilityIsAllowed(name, vulnerabilities, visiting = new Set()) {
  if (visiting.has(name)) return false;
  if (!Object.hasOwn(vulnerabilities, name)) return false;

  const vulnerability = vulnerabilities[name];
  if (!isPlainObject(vulnerability)) return false;
  if (!Array.isArray(vulnerability.via) || vulnerability.via.length === 0) return false;

  const nextVisiting = new Set(visiting).add(name);
  return vulnerability.via.every((via) => {
    if (typeof via === 'string') {
      return vulnerabilityIsAllowed(via, vulnerabilities, nextVisiting);
    }
    return advisoryIsAllowed(via);
  });
}

export function evaluateAudit(report) {
  if (!isPlainObject(report) || !isPlainObject(report.vulnerabilities)) {
    return { ok: false, accepted: [], blockers: ['invalid npm audit report'] };
  }

  const accepted = [];
  const blockers = [];

  for (const [name, vulnerability] of Object.entries(report.vulnerabilities)) {
    // An entry that is not a well-formed advisory record cannot be shown to be
    // below the blocking severities, so it blocks rather than being skipped.
    if (!isPlainObject(vulnerability) || typeof vulnerability.severity !== 'string') {
      blockers.push(name);
      continue;
    }

    if (!BLOCKING_SEVERITIES.has(vulnerability.severity)) continue;

    if (vulnerabilityIsAllowed(name, report.vulnerabilities)) {
      accepted.push(name);
    } else {
      blockers.push(name);
    }
  }

  return { ok: blockers.length === 0, accepted, blockers };
}

function run() {
  const boundaryViolations = evaluateRouterBoundary(sourceFiles(join(process.cwd(), 'src')));
  if (boundaryViolations.length > 0) {
    process.stderr.write(`Router risk-acceptance boundary violated:\n${boundaryViolations.join('\n')}\n`);
    process.exit(1);
  }

  const result = spawnSync('npm', ['audit', '--json', '--audit-level=high'], {
    cwd: process.cwd(),
    encoding: 'utf8',
    maxBuffer: AUDIT_MAX_BUFFER_BYTES,
  });

  // A nonzero npm exit status is expected whenever advisories are present, so
  // the classification - not the status - decides whether a report exists.
  const outcome = classifyAuditResult(result);
  if (outcome.kind !== 'report') {
    process.stderr.write(`npm audit gate failed closed [${outcome.kind}]: ${outcome.detail}\n`);
    process.exit(1);
  }

  const evaluation = evaluateAudit(outcome.report);
  if (!evaluation.ok) {
    process.stderr.write(
      `Blocking high/critical npm advisories: ${evaluation.blockers.join(', ')}\n`
    );
    process.exit(1);
  }

  if (evaluation.accepted.length > 0) {
    process.stdout.write(
      `Accepted GHSA-qwww-vcr4-c8h2 only for the non-RSC BrowserRouter SPA: ${evaluation.accepted.join(', ')}\n`
    );
  } else {
    process.stdout.write('No high or critical npm advisories found.\n');
  }
}

if (process.argv[1] && fileURLToPath(import.meta.url) === process.argv[1]) {
  run();
}

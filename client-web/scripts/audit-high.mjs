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

function vulnerabilityIsAllowed(name, vulnerabilities, visiting = new Set()) {
  if (visiting.has(name)) return false;

  const vulnerability = vulnerabilities[name];
  if (!vulnerability || !Array.isArray(vulnerability.via) || vulnerability.via.length === 0) {
    return false;
  }

  const nextVisiting = new Set(visiting).add(name);
  return vulnerability.via.every((via) => {
    if (typeof via === 'string') {
      return vulnerabilityIsAllowed(via, vulnerabilities, nextVisiting);
    }
    return advisoryIsAllowed(via);
  });
}

export function evaluateAudit(report) {
  if (!report || typeof report !== 'object' || !report.vulnerabilities) {
    return { ok: false, accepted: [], blockers: ['invalid npm audit report'] };
  }

  const blocking = Object.entries(report.vulnerabilities).filter(([, vulnerability]) =>
    BLOCKING_SEVERITIES.has(vulnerability?.severity)
  );
  const accepted = [];
  const blockers = [];

  for (const [name] of blocking) {
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
  });

  let report;
  try {
    report = JSON.parse(result.stdout);
  } catch {
    process.stderr.write(result.stderr || 'npm audit did not return valid JSON\n');
    process.exit(1);
  }

  const evaluation = evaluateAudit(report);
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

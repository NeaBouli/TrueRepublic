import { existsSync, readFileSync, readdirSync } from 'node:fs';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { gzipSync } from 'node:zlib';

export const BUNDLE_BUDGET = Object.freeze({
  routeEntries: 19,
  entry: Object.freeze({ raw: 260_000, gzip: 85_000 }),
  route: Object.freeze({ raw: 25_000, gzip: 7_000 }),
  chunk: Object.freeze({ raw: 1_100_000, gzip: 150_000 }),
  total: Object.freeze({ raw: 1_900_000, gzip: 390_000 }),
});

export function evaluateBundleBudget(manifest, files, budget = BUNDLE_BUDGET) {
  const violations = [];
  const entries = Object.values(manifest).filter((item) => item.isEntry === true);
  const routeEntries = Object.entries(manifest).filter(
    ([source, item]) =>
      source.startsWith('src/components/') &&
      source.endsWith('.tsx') &&
      item.isDynamicEntry === true
  );

  if (entries.length !== 1) {
    violations.push(`expected exactly 1 application entry, found ${entries.length}`);
  }
  if (routeEntries.length !== budget.routeEntries) {
    violations.push(
      `expected ${budget.routeEntries} lazy route entries, found ${routeEntries.length}`
    );
  }

  const check = (label, size, limit) => {
    if (!size) {
      violations.push(`${label}: emitted file is missing`);
      return;
    }
    if (size.raw > limit.raw) {
      violations.push(`${label}: ${size.raw} raw bytes exceeds ${limit.raw}`);
    }
    if (size.gzip > limit.gzip) {
      violations.push(`${label}: ${size.gzip} gzip bytes exceeds ${limit.gzip}`);
    }
  };

  for (const entry of entries) check('entry', files[entry.file], budget.entry);
  for (const [source, route] of routeEntries) {
    check(`route ${source}`, files[route.file], budget.route);
  }

  const javascript = Object.entries(files).filter(([file]) => file.endsWith('.js'));
  for (const [file, size] of javascript) check(`chunk ${file}`, size, budget.chunk);

  const total = javascript.reduce(
    (sum, [, size]) => ({ raw: sum.raw + size.raw, gzip: sum.gzip + size.gzip }),
    { raw: 0, gzip: 0 }
  );
  check('total JavaScript', total, budget.total);

  return {
    ok: violations.length === 0,
    violations,
    measurements: {
      entry: entries.length === 1 ? files[entries[0].file] : null,
      routeEntries: routeEntries.length,
      maxRoute: routeEntries.reduce(
        (max, [, route]) => Math.max(max, files[route.file]?.gzip ?? 0),
        0
      ),
      total,
    },
  };
}

function readBuild(buildDirectory) {
  const manifestPath = join(buildDirectory, '.vite', 'manifest.json');
  if (!existsSync(manifestPath)) {
    throw new Error(`Vite manifest not found: ${manifestPath}`);
  }

  const manifest = JSON.parse(readFileSync(manifestPath, 'utf8'));
  const assetsDirectory = join(buildDirectory, 'assets');
  const files = {};
  for (const file of readdirSync(assetsDirectory)) {
    const contents = readFileSync(join(assetsDirectory, file));
    files[`assets/${file}`] = {
      raw: contents.byteLength,
      gzip: gzipSync(contents, { level: 9 }).byteLength,
    };
  }
  return { manifest, files };
}

function run() {
  let build;
  try {
    build = readBuild(join(process.cwd(), 'dist'));
  } catch (error) {
    process.stderr.write(`${error instanceof Error ? error.message : String(error)}\n`);
    process.exit(1);
  }

  const result = evaluateBundleBudget(build.manifest, build.files);
  if (!result.ok) {
    process.stderr.write(`Bundle budget failed:\n${result.violations.join('\n')}\n`);
    process.exit(1);
  }

  const { entry, routeEntries, maxRoute, total } = result.measurements;
  process.stdout.write(
    `Bundle budget passed: entry ${entry.gzip} gzip bytes; ` +
      `${routeEntries} lazy routes; max route ${maxRoute} gzip bytes; ` +
      `total JavaScript ${total.gzip} gzip bytes.\n`
  );
}

if (process.argv[1] === fileURLToPath(import.meta.url)) run();

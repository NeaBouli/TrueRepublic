#!/usr/bin/env node
// Deterministic static contract audit for the GH-328 landing architecture
// visualization in docs/index.html. Dependency-free; reads only repository
// files and exits non-zero on the first run with any failed assertion.
//
// Usage: node scripts/landing-architecture-contract-audit.mjs

import { existsSync, readFileSync } from "node:fs";
import { dirname, join, normalize, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const repoRoot = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const landingPath = join(repoRoot, "docs", "index.html");
const statusPath = join(repoRoot, "docs", "status.json");

const html = readFileSync(landingPath, "utf8");
const status = JSON.parse(readFileSync(statusPath, "utf8"));

const failures = [];
let passed = 0;

function check(label, condition) {
  if (condition) {
    passed += 1;
    console.log(`  OK   ${label}`);
  } else {
    failures.push(label);
    console.log(`  FAIL ${label}`);
  }
}

function textOf(fragment) {
  return fragment
    .replace(/<[^>]+>/g, " ")
    .replace(/&amp;/g, "&")
    .replace(/&mdash;/g, "-")
    .replace(/&[#a-zA-Z0-9]+;/g, " ")
    .replace(/\s+/g, " ")
    .trim();
}

function cssRule(selector) {
  const escaped = selector.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const match = html.match(new RegExp(`(^|[\\s}])${escaped}\\s*\\{([^}]*)\\}`, "m"));
  return match ? match[2] : "";
}

// --- Section extraction -----------------------------------------------------
const sectionMatch = html.match(
  /<section id="architecture"[^>]*>([\s\S]*?)<\/section>/,
);
check("architecture section exists", Boolean(sectionMatch));
if (!sectionMatch) {
  console.log(`\nFAILED: ${failures.length} landing architecture contract failure(s)`);
  process.exit(1);
}
const sectionOpen = sectionMatch[0].slice(0, sectionMatch[0].indexOf(">") + 1);
const section = sectionMatch[1];
const sectionText = textOf(section);

// --- 1. Four delivery strands -------------------------------------------------
const strandHeadings = [...section.matchAll(/<article class="arch-strand"[\s\S]*?<h3 id="([^"]+)">([\s\S]*?)<\/h3>/g)]
  .map((m) => ({ id: m[1], text: textOf(m[2]) }));
check(
  "exactly four delivery strands in order: Design, Implementation, Simulation & Verification, Release",
  JSON.stringify(strandHeadings.map((h) => h.text)) ===
    JSON.stringify(["Design", "Implementation", "Simulation & Verification", "Release"]),
);
for (const heading of strandHeadings) {
  check(`strand "${heading.text}" is labelled by its heading`, section.includes(`aria-labelledby="${heading.id}"`));
}

// --- 2. Status vocabulary: text and color ------------------------------------
const statuses = [
  ["implemented", "Implemented"],
  ["partial", "Partial"],
  ["test", "Test-only"],
  ["quarantined", "Quarantined"],
  ["deferred", "Deferred"],
];
const legend = (section.match(/<ul class="arch-legend"[\s\S]*?<\/ul>/) || [""])[0];
for (const [key, label] of statuses) {
  const badge = new RegExp(`<span class="arch-status arch-status-${key}">[\\s\\S]*?</span> ${label}</span>`);
  check(`legend shows "${label}" as visible text`, badge.test(legend));
  const strandUses = [...section.matchAll(new RegExp(`class="arch-status arch-status-${key}"`, "g"))].length;
  check(`"${label}" is used on at least one strand node`, strandUses >= 2);
  const rule = cssRule(`.arch-status-${key}`);
  check(`"${label}" has its own color rule`, /color:\s*#[0-9a-fA-F]{6}/.test(rule) && /background:/.test(rule));
  check(`"${label}" has a non-color border-style cue`, /border-style:\s*(solid|double|dashed|dotted)/.test(rule));
}
const borderStyles = statuses.map(([key]) => (cssRule(`.arch-status-${key}`).match(/border-style:\s*(\w+)/) || [])[1]);
check(
  "quarantined, deferred and test-only badges use distinct border styles",
  new Set(borderStyles.slice(2)).size === 3,
);
const everyBadgeLabelled = [...section.matchAll(/<span class="arch-status arch-status-(\w+)">([\s\S]*?)<\/span> ([^<]+)<\/span>/g)]
  .every((m) => statuses.some(([key, label]) => key === m[1] && m[3].trim() === label));
check("every status badge carries its matching text label", everyBadgeLabelled);
check(
  "decorative status glyphs are hidden from assistive technology",
  [...section.matchAll(/<span class="arch-status [^"]+"><span([^>]*)>/g)].every((m) => m[1].includes('aria-hidden="true"')),
);

// --- 3. Strict delivery rule ------------------------------------------------
const flow = (section.match(/<ol class="arch-flow"[\s\S]*?<\/ol>/) || [""])[0];
const steps = [...flow.matchAll(/<li><strong>([\s\S]*?)<\/strong>/g)].map((m) => textOf(m[1]));
check(
  "delivery rule is an ordered list: boundary -> one owned node -> verification -> integration -> release gate",
  JSON.stringify(steps) ===
    JSON.stringify([
      "Architecture boundary",
      "One owned implementation node",
      "Verification",
      "Integration",
      "Release gate",
    ]),
);

// --- 4. Relative links to authoritative documents ----------------------------
const hrefs = [...section.matchAll(/href="([^"]*)"/g)].map((m) => m[1]);
for (const target of ["architecture/MAP.md", "architecture/EXECUTION_PLAN.md"]) {
  check(`links to ${target}`, hrefs.includes(target));
  const resolved = normalize(join(repoRoot, "docs", target));
  check(`${target} resolves to an existing file under docs/`, resolved.startsWith(join(repoRoot, "docs")) && existsSync(resolved));
}
check(
  "every section link is a relative in-repository docs URL",
  hrefs.length > 0 && hrefs.every((href) => /^[A-Za-z0-9_\-/]+\.md$/.test(href) && !href.startsWith("/")),
);
const linkLabels = [...section.matchAll(/<a [^>]*>([\s\S]*?)<\/a>/g)].map((m) => textOf(m[1]));
check(
  "link labels are meaningful (no generic 'click here'/'read more'/'here')",
  linkLabels.length === 2 && linkLabels.every((l) => l.length >= 12 && !/^(click here|read more|here|more|link)$/i.test(l)),
);

// --- 5. Canonical rollout and production readiness --------------------------
const canonical = `${status.rollout.completed}/${status.rollout.total}`;
check("docs/status.json keeps canonical rollout 36/59", canonical === "36/59");
check("docs/status.json keeps rollout.production_ready false", status.rollout.production_ready === false);
check(`section shows canonical rollout ${canonical}`, sectionText.includes(canonical));
const fractions = [...sectionText.matchAll(/\b(\d+)\/59\b/g)].map((m) => m[0]);
check("section shows no other N/59 rollout value", fractions.every((f) => f === canonical));
check("section shows production readiness false", /<span class="arch-baseline-value">false<\/span>/.test(section));
check("section states that no rollout credit is earned", /earn(s)? no rollout credit/.test(sectionText));
check(
  "section makes no production-ready or completion claim",
  !/production[- ]ready\b|\bapproved for production|audit (is )?(complete|closed)|rollout (is )?complete/i.test(sectionText),
);

// --- 6. Deferred tracks -------------------------------------------------------
check(
  "states V3 ballot privacy and V4 edge work are deferred and outside the Basic rollout",
  /V3 ballot privacy[^.]*and V4 edge work are deferred and outside the Basic rollout/.test(sectionText),
);

// --- 7. No internal or private detail --------------------------------------
const forbiddenPublic = [
  [/agent-bridge|BRIDGE\.md|\.fleet\b|PROJECT_STATE|SECURITY_NOTES|ACTION_LOG/i, "internal bridge paths"],
  [/\b(Kimi|Codex|Claude|Grok|Sol|jev)\b|\bworker\b|\blane [A-Z]\b/i, "worker routing"],
  [/hetzner|sandbox|operator-a|ssh\b|\b\d{1,3}(\.\d{1,3}){3}\b|localhost/i, "hostnames or addresses"],
  [/TRR-\d+|GO-20\d\d-\d+|RUSTSEC|CVE-|CWA-|\bG[1-7]\b|\bC[1-4]\b/, "security finding identifiers"],
  [/\/Users\/|~\/|worktree/i, "private filesystem or planning detail"],
];
for (const [pattern, label] of forbiddenPublic) {
  check(`section exposes no ${label}`, !pattern.test(section));
}

// --- 8. No remote or runtime additions -------------------------------------
check("page contains no <script> element", !/<script\b/i.test(html));
check("page contains no inline event handler attribute", !/\son[a-z]+\s*=/i.test(html));
check("page contains no javascript: URL", !/javascript:/i.test(html));
check("page loads no external stylesheet or font", !/<link[^>]+rel="stylesheet"|@import|@font-face|fonts\.(googleapis|gstatic)/i.test(html));
check("page references no CDN or telemetry host", !/cdn\.|jsdelivr|unpkg|cdnjs|googletagmanager|google-analytics|plausible|matomo|segment\.io/i.test(html));
check("section contains no remote URL", !/(https?:)?\/\//i.test(section));
check("section embeds no iframe, object, embed or remote media", !/<(iframe|object|embed|img|video|audio|source)\b/i.test(section));
check("section uses no inline style attribute", !/\sstyle="/i.test(section));

// --- 9. Static layout and accessibility contracts --------------------------
const headingOrder = [...section.matchAll(/<h([1-6])\b/g)].map((m) => Number(m[1]));
check("section begins with an h2 and nests only h3 headings", headingOrder[0] === 2 && headingOrder.slice(1).every((h) => h === 3));
check(
  "section region is labelled by its visible heading",
  /aria-labelledby="architecture-title"/.test(sectionOpen) && section.includes('id="architecture-title"'),
);
const actionRule = cssRule(".arch-actions .btn");
check("action links have min-height 44px", /min-height:\s*44px/.test(actionRule));
check("action links have min-width 44px", /min-width:\s*44px/.test(actionRule));
check("action links are inline-flex targets", /display:\s*inline-flex/.test(actionRule));
check("action links wrap instead of overflowing", /flex-wrap:\s*wrap/.test(cssRule(".arch-actions")));
const focusRule = cssRule(".architecture a:focus-visible");
check("links in the section have a visible 3px focus outline", /outline:\s*3px solid/.test(focusRule) && /outline-offset:\s*\d+px/.test(focusRule));
check("strand grid collapses responsively", /repeat\(auto-fit,\s*minmax\(250px,\s*1fr\)\)/.test(cssRule(".arch-strands")));
check("strand cards may shrink below content width", /min-width:\s*0/.test(cssRule(".arch-strand")));
check(
  "delivery flow stacks to one column at or below 1000px",
  /@media \(max-width: 1000px\)\s*\{\s*\.arch-flow\s*\{\s*grid-template-columns:\s*minmax\(0,\s*1fr\)/.test(html),
);
check(
  "flow arrows are decorative (empty alternative text)",
  /content:\s*"\\2192"\s*\/\s*""/.test(html) && /content:\s*"\\2193"\s*\/\s*""/.test(html),
);
check("viewport meta allows responsive scaling", /<meta name="viewport" content="width=device-width, initial-scale=1\.0">/.test(html));

// --- Result ------------------------------------------------------------------
if (failures.length > 0) {
  console.log(`\nFAILED: ${failures.length} of ${passed + failures.length} landing architecture contract assertions`);
  process.exit(1);
}
console.log(`\nPASSED: ${passed} landing architecture contract assertions`);

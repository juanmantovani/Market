#!/usr/bin/env node
// Enforces the backend coverage threshold of AGENTS.md over the profile
// `pnpm test:backend:coverage` writes. `go tool cover -func` only prints
// numbers; this is what makes a drop fail the build.

import { readFileSync } from 'node:fs';

const PROFILE = 'apps/backend/coverage.out';
const MODULE_PREFIX = 'market/backend/';
const THRESHOLD = 80;

function readProfile(path) {
  let contents;
  try {
    contents = readFileSync(path, 'utf8');
  } catch {
    console.error(`coverage profile not found at ${path}: run pnpm test:backend:coverage first`);
    process.exit(1);
  }

  const packages = new Map();
  for (const line of contents.split('\n').slice(1)) {
    const match = line.match(/^(.*):\d+\.\d+,\d+\.\d+ (\d+) (\d+)$/);
    if (!match) {
      continue;
    }

    const [, file, statements, count] = match;
    const name = file.slice(0, file.lastIndexOf('/')).replace(MODULE_PREFIX, '');
    const totals = packages.get(name) ?? { statements: 0, covered: 0 };
    totals.statements += Number(statements);
    if (Number(count) > 0) {
      totals.covered += Number(statements);
    }
    packages.set(name, totals);
  }

  return packages;
}

const packages = readProfile(PROFILE);
if (packages.size === 0) {
  console.error(`${PROFILE} holds no coverage data`);
  process.exit(1);
}

const failures = [];
let statements = 0;
let covered = 0;

for (const [name, totals] of [...packages].sort()) {
  const percentage = (totals.covered / totals.statements) * 100;
  statements += totals.statements;
  covered += totals.covered;

  if (percentage < THRESHOLD) {
    failures.push(`${name}: ${percentage.toFixed(1)}% of statements, want at least ${THRESHOLD}%`);
    console.log(`  FAIL  ${percentage.toFixed(1).padStart(5)}%  ${name}`);
    continue;
  }
  console.log(`  ok    ${percentage.toFixed(1).padStart(5)}%  ${name}`);
}

const total = (covered / statements) * 100;
console.log(`\n  total ${total.toFixed(1)}% of statements over the measured packages`);

if (total < THRESHOLD) {
  failures.push(`total: ${total.toFixed(1)}% of statements, want at least ${THRESHOLD}%`);
}

if (failures.length > 0) {
  console.error(`\nbackend coverage below the threshold:\n  ${failures.join('\n  ')}`);
  process.exit(1);
}

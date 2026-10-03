import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const ROOT_DIR = path.resolve(__dirname, '..');

const packageJsonPath = path.join(ROOT_DIR, 'package.json');
const packageLockPath = path.join(ROOT_DIR, 'package-lock.json');

if (!fs.existsSync(packageJsonPath) || !fs.existsSync(packageLockPath)) {
  console.error('package.json or package-lock.json missing.');
  process.exit(1);
}

const packageJson = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));
const packageLock = JSON.parse(fs.readFileSync(packageLockPath, 'utf8'));

const allDependencies = {
  ...(packageJson.dependencies || {}),
  ...(packageJson.devDependencies || {}),
};

async function checkDeprecated() {
  let hasDeprecated = false;

  for (const pkg of Object.keys(allDependencies)) {
    // Check lockfile for the installed version
    const lockEntry = packageLock.packages[`node_modules/${pkg}`];
    if (!lockEntry) {
      console.warn(`Could not find ${pkg} in package-lock.json`);
      continue;
    }
    const version = lockEntry.version;
    if (!version) continue;

    try {
      const url = `https://registry.npmjs.org/${pkg}/${version}`;
      const res = await fetch(url);
      if (!res.ok) {
        console.warn(`Failed to fetch metadata for ${pkg}@${version}`);
        continue;
      }
      const data = await res.json();
      if (data.deprecated) {
        console.error(`DEPRECATED: ${pkg}@${version}`);
        console.error(`Reason: ${data.deprecated}`);
        hasDeprecated = true;
      }
    } catch (err) {
      console.error(`Error checking ${pkg}@${version}: ${err.message}`);
    }
  }

  if (hasDeprecated) {
    console.error('One or more dependencies are deprecated. Please update or replace them.');
    process.exit(1);
  }
  
  console.log('All dependencies checked, no deprecated packages found.');
}

checkDeprecated().catch(err => {
  console.error('Fatal error:', err);
  process.exit(1);
});

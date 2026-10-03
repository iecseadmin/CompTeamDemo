import fs from 'fs';
import path from 'path';

const MEMBERS_FILE = path.resolve('src', 'data', 'members.json');
const RATINGS_FILE = path.resolve('src', 'data', 'ratings.json');
const API_BASE = 'https://cp-rating-api.vercel.app';

// Validate member status
const ALLOWED_STATUSES = ['Junior Team', 'Senior Team', 'Retired', 'Board Team Member'];

async function fetchWithRetry(url, retries = 3) {
  for (let i = 0; i < retries; i++) {
    try {
      const controller = new AbortController();
      const id = setTimeout(() => controller.abort(), 10000);
      const res = await fetch(url, { signal: controller.signal });
      clearTimeout(id);
      
      if (!res.ok) {
        if (res.status === 404) return { status: 'not-found', rating: null };
        throw new Error(`HTTP ${res.status}`);
      }
      const data = await res.json();
      return { status: 'ok', data };
    } catch (err) {
      if (i === retries - 1) {
        return { status: 'error', error: err.message, rating: null };
      }
      // wait before retry
      await new Promise(r => setTimeout(r, 1000 * (i + 1)));
    }
  }
}

async function fetchPlatform(platform, username) {
  if (!username) return { status: 'ok', rating: null, username: null };
  
  const url = `${API_BASE}/${platform}/${username}`;
  const result = await fetchWithRetry(url);
  
  if (result.status === 'ok') {
    return { status: 'ok', rating: result.data.rating || null, username };
  }
  return { status: result.status, rating: null, username };
}

async function main() {
  console.log('Reading members.json...');
  if (!fs.existsSync(MEMBERS_FILE)) {
    console.error('members.json not found!');
    process.exit(1);
  }
  
  const members = JSON.parse(fs.readFileSync(MEMBERS_FILE, 'utf-8'));
  const oldRatings = fs.existsSync(RATINGS_FILE) ? JSON.parse(fs.readFileSync(RATINGS_FILE, 'utf-8')) : [];
  const oldRatingsMap = new Map(oldRatings.map(r => [r.name, r]));

  const results = [];
  const now = new Date().toISOString();

  for (const member of members) {
    if (!ALLOWED_STATUSES.includes(member.memberStatus)) {
      console.error(`Invalid memberStatus for ${member.name}: ${member.memberStatus}. Must be one of ${ALLOWED_STATUSES.join(', ')}`);
      process.exit(1);
    }
    
    console.log(`Fetching stats for ${member.name}...`);
    
    // Fetch platforms exclusively using the public API
    const [leetcode, codeforces, codechef, atcoder] = await Promise.all([
      fetchPlatform('leetcode', member.leetcodeUsername),
      fetchPlatform('codeforces', member.codeforcesUsername),
      fetchPlatform('codechef', member.codechefUsername),
      fetchPlatform('atcoder', member.atcoderUsername)
    ]);
    
    // Preserve old data on transient error
    const old = oldRatingsMap.get(member.name);
    
    const resolvePlatform = (current, oldPlatform) => {
      if (current.status === 'error' && oldPlatform && oldPlatform.status === 'ok') {
        console.log(`  Warning: Failed to fetch for ${current.username}, using previous successful rating.`);
        return { ...oldPlatform, _transientError: true };
      }
      return current;
    };

    const finalLeetcode = resolvePlatform(leetcode, old?.leetcode);
    const finalCodeforces = resolvePlatform(codeforces, old?.codeforces);
    const finalCodechef = resolvePlatform(codechef, old?.codechef);
    const finalAtcoder = resolvePlatform(atcoder, old?.atcoder);

    results.push({
      name: member.name,
      memberStatus: member.memberStatus,
      leetcode: finalLeetcode,
      codeforces: finalCodeforces,
      codechef: finalCodechef,
      atcoder: finalAtcoder,
      updatedAt: now
    });
  }

  // Sort by Name (Alphabetical)
  results.sort((a, b) => a.name.localeCompare(b.name));

  fs.writeFileSync(RATINGS_FILE, JSON.stringify(results, null, 2), 'utf-8');
  console.log('Done writing ratings.json');
}

main().catch(err => {
  console.error('Fatal error:', err);
  process.exit(1);
});

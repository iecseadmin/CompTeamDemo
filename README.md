# CompTeam — Competitive Programming Statistics

This is a production-ready, static website for a competitive programming team. It displays "Online Stats" (live ratings from LeetCode, Codeforces, CodeChef, and AtCoder) and "Offline Stats" (competition and event achievements). The design is sparse, terminal-inspired, and strictly focused on data readability.

The site uses React, TypeScript, and Vite, and is designed to be deployed to GitHub Pages. It relies on the [CP Rating API](https://cp-rating-api.vercel.app/) to fetch public ratings without requiring credentials.

## What a Maintainer Edits

All user-facing content is controlled via JSON files in `src/data/`:

1. **`members.csv`**: Edit this file with team member details, then run the Go tool to generate the roster.
2. **`src/data/offline-events.json`**: Edit this file manually to add historical achievements and event records.

**Do not manually edit `src/data/ratings.json`**. This file is generated automatically by the daily rating refresh script.

## Importing Team Members

To update the canonical team roster (`src/data/members.json`), edit `members.csv` with the following headers:
`Name,LeetCode Username,CodeChef Username,Codeforces Username,AtCoder Username,Member Status`

Then run the Go utility using Windows PowerShell:
```powershell
go run .\tools\csv-to-members-json\main.go -input .\members.csv -output .\src\data\members.json
```

- **Permitted Member Statuses**: `Junior Team`, `Senior Team`, `Retired`, or `Board Team Member`.
- **Missing Accounts**: Leave the cell blank in the CSV. The Go tool will convert it to `null`. In the UI, this displays as a grey, non-clickable `N/A`.

## Ratings

Ratings are refreshed automatically every day. 
- **Missing Ratings**: If a username is `null` or if a platform temporarily fails to return data, the interface will display an unclickable `N/A`. Transient API failures will safely fall back to the last known successful rating for that platform.
- **Sorting Rules**: The Online Stats table defaults to alphabetical sorting by Name. You can click any column header to sort by a specific platform rating.

## Local Development (Windows PowerShell)

All commands must be executed in Windows PowerShell from the project root.

**First-time setup:**
```powershell
npm ci
```

**Validate JSON data schemas:**
```powershell
npm run validate:data
```

**Run local development server:**
```powershell
npm run dev
```

**Refresh ratings manually:**
```powershell
npm run fetch:ratings
```

**Build for production:**
```powershell
npm run build
```

## Deployment and Automation

This project includes two GitHub Actions workflows located in `.github/workflows/`:
1. **`deploy-pages.yml`**: Triggers on pushes to the default branch. It validates data, builds the site, and deploys it.
2. **`refresh-stats.yml`**: Runs automatically every day at 02:00 UTC. It fetches the latest ratings via `npm run fetch:ratings`, builds the site with the new data, and deploys it directly to GitHub Pages. It **does not** commit the generated JSON back into your Git repository history.

## Pre-Publishing Checklist

Before publishing this site publicly, you must:
1. **Rename the repository** to `CompTeam` (or update `vite.config.ts` if you choose a different name). The current Vite configuration assumes the site is hosted at `https://<your-username>.github.io/CompTeam/`.
2. **Replace the roster placeholders** in `members.csv` with real team members, and run the Go script.
3. **Replace the event placeholders** in `src/data/offline-events.json` with your real history.
4. **Enable GitHub Pages**: Go to your repository Settings → Pages. Under "Build and deployment", set the Source to "GitHub Actions".

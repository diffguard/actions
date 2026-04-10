# DiffGuard Actions

Open-source GitHub Actions used by [DiffGuard](https://www.diffguard.io) — an organizational dependency intelligence platform for software supply chain security.

This repository is the public, auditable half of DiffGuard's open-source boundary. The Actions ship here so that customers can read every line of code that runs inside their CI pipelines.

## Contents

| Directory | What it is |
|---|---|
| `active-gate/` | GitHub Action — submits a lockfile diff at PR time for security analysis |
| `passive-monitor/` | GitHub Action — submits resolved lockfile state after install |
| `action-monitor/` | GitHub Action — parses workflow `uses:` references and submits them for analysis |

## Using the Actions

```yaml
- uses: diffguard/actions/active-gate@v0.1.0
  with:
    api-token: ${{ secrets.DIFFGUARD_TOKEN }}
```

## License

MIT — see `LICENSE`.

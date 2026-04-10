# DiffGuard Actions

Open-source GitHub Actions and lockfile parsers used by [DiffGuard](https://github.com/diffguard/DiffGuard) — an organizational dependency intelligence platform for software supply chain security.

This repository is the public, auditable half of DiffGuard's open-source boundary. The Actions and the lockfile parsers ship here so that customers can read every line of code that runs inside their CI pipelines.

## Contents

| Directory | What it is |
|---|---|
| `active-gate/` | GitHub Action — submits a lockfile diff at PR time for security analysis |
| `passive-monitor/` | GitHub Action — submits resolved lockfile state after install |
| `action-monitor/` | GitHub Action — parses workflow `uses:` references and submits them for analysis |
| `shared/lockfile/` | Lockfile parsers (npm, yarn, pnpm, poetry, pip-tools). Pure Go, no CGO. |

## Using the Actions

```yaml
- uses: diffguard/actions/active-gate@v0.1.0
  with:
    api-token: ${{ secrets.DIFFGUARD_TOKEN }}
```

## Using the lockfile parsers as a Go library

```go
import "github.com/diffguard/actions/shared/lockfile"
```

## License

MIT — see `LICENSE`.

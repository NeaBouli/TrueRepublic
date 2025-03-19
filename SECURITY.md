# TrueRepublic - CI/CD, Security & Bug Bounty

## Continuous Integration (CI) & Deployment (CD) for TrueRepublic

### Continuous Integration (CI)
CI ensures that every code change triggers automated tests and checks to catch errors early and keep the codebase stable.

#### 1. CI with GitHub Actions
- Automated tests for:
  - Blockchain (Go): `blockchain/.github/workflows/go-ci.yml`
  - Smart Contracts (Rust): `contracts/.github/workflows/rust-ci.yml`
  - Web Wallet (React): `web-wallet/.github/workflows/react-ci.yml`
  - Mobile Wallet (React Native): `mobile-wallet/.github/workflows/react-native-ci.yml`
- Automated builds & error checking

#### 2. Continuous Deployment (CD)
- Automated deployments for:
  - Blockchain: `blockchain/.github/workflows/blockchain-cd.yml`
  - Smart Contracts: `contracts/.github/workflows/contracts-cd.yml`
  - Web Wallet: `web-wallet/.github/workflows/web-cd.yml`
  - Mobile Wallet: `mobile-wallet/.github/workflows/mobile-cd.yml`

### Security Measures (Audits & API Monitoring)

#### 1. Smart Contract Security Audits
- Tools: `cargo-audit`, `clippy`
- Commands: 
  ```bash
  cargo install cargo-audit cargo-clippy
  cargo clippy --all-targets --all-features
  cargo audit


#### 2. Automated Dependency Scans
- Workflow: `.github/workflows/dependency-check.yml`
- Runs daily at 03:00 UTC and on every push to `main`.

#### 3. API Monitoring with Prometheus & Grafana
- Setup:
  ```bash
  docker run -d --name prometheus -p 9090:9090 prom/prometheus
  docker run -d --name grafana -p 3000:3000 grafana/grafana

- Integration in `blockchain/app.go`:
  ```go
  import (
      "github.com/prometheus/client_golang/prometheus"
      "github.com/prometheus/client_golang/prometheus/promhttp"
      "net/http"
  )

  var apiRequests = prometheus.NewCounterVec(
      prometheus.CounterOpts{
          Name: "api_requests_total",
          Help: "Total API requests",
      },
      []string{"method"},
  )

  func main() {
      prometheus.MustRegister(apiRequests)
      http.Handle("/metrics", promhttp.Handler())
      go http.ListenAndServe(":8080", nil)
  }


#### 4. Firewall & DDoS Protection
- UFW setup:
  ```bash
  sudo ufw allow 26656  # Tendermint P2P
  sudo ufw allow 26657  # RPC-Server
  sudo ufw allow 9090   # Prometheus Monitoring
  sudo ufw allow 443/tcp  # HTTPS for API
  sudo ufw enable

Cloudflare: Enable "Under Attack Mode" and rate-limiting (max. 100 API requests/minute).

### Emergency Backup & Recovery System

#### 1. Blockchain Backup
- CronJob (daily at 03:00):
  ```bash
  0 3 * * * tar -czf ~/backup/truerepublic_$(date +\%F).tar.gz ~/.truerepublic

Remote backup script: scripts/backup.sh
bash

#!/bin/bash
tar -czf ~/backup/truerepublic_$(date +%F).tar.gz ~/.truerepublic
rclone copy ~/backup/truerepublic_$(date +%F).tar.gz remote:TrueRepublicBackups

CronJob for script:
bash

0 4 * * * ~/backup/backup.sh

- Recovery:
  ```bash
  rclone copy remote:TrueRepublicBackups/truerepublic_LATEST.tar.gz ~/
  tar -xzf truerepublic_LATEST.tar.gz -C ~/
  truerepublicd start


#### 2. API Server Backup
- PM2 setup:
  ```bash
  npm install pm2 -g
  pm2 start api-server.js --name truerepublic-api
  pm2 save
  pm2 startup

Database backup (daily at 02:00):
bash

0 2 * * * pg_dump truerepublic_db > ~/backup/api_backup_$(date +\%F).sql

Recovery:
bash

psql truerepublic_db < ~/backup/api_backup_LATEST.sql


#### 3. Web & Mobile Wallet Backup
- Web Wallet rollback (Vercel):
  ```bash
  npx vercel rollback --to=previous --token=${{ secrets.VERCEL_TOKEN }}

Mobile Wallet rollback: Use Google Play Console or App Store Connect to revert to the previous version.

#### 4. Monitoring
- UptimeRobot: Add `https://api.truerepublic.network` with Telegram alerts.

### Bug Bounty Program
- Page: `bug-bounty/pages/index.js` (Next.js, deployable with `npx vercel --prod --token=${{ secrets.VERCEL_TOKEN }}`)
- GitHub Issues: `.github/ISSUE_TEMPLATE/bug_bounty.md`
- Optional platforms: HackerOne, Immunefi

### Community Governance
- Smart Contracts for voting: `contracts/governance.rs`
- Community voting with PNYX stakes
- On-chain proposal mechanism

TrueRepublic is now fully automated, secure, and community-driven!

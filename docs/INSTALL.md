# TrueRepublic installieren

> Recovery-Status: Die folgenden Schritte sind für lokale Entwicklung und
> Recovery-Testnets bestimmt. Das Projekt ist nicht für Mainnet oder reale
> Gelder freigegeben.

## Voraussetzungen

- Go 1.26.5
- Git
- optional: Rust 1.75+ für `contracts/`
- optional: Node.js 22+ und npm für `client-web/`
- optional: Docker 24+ mit Compose v2

Die veralteten Verzeichnisse `web-wallet` und `mobile-wallet` sind keine
Installationsziele für reale Schlüssel.

## Quellcode und Tests

```bash
git clone https://github.com/NeaBouli/TrueRepublic.git
cd TrueRepublic
git switch --track origin/fix/GH-21-node-lifecycle
go build -o build/truerepublicd .
go test ./... -count=1
```

Der aktuelle Recovery-Stack liegt bis zur geordneten Zusammenführung in den
Draft-PRs #9, #15, #16, #17, #18, #19, #22, #23 und #24. `main` enthält diese
Änderungen noch nicht vollständig. Der explizite Branch-Wechsel ist deshalb
für die folgenden Recovery-Node-Befehle erforderlich; er ist keine
Mainnet-Freigabe.

## Lokalen Recovery-Node starten

```bash
./build/truerepublicd init local-node \
  --chain-id truerepublic-local-1 \
  --home "$HOME/.truerepublic"

./build/truerepublicd keys add local-user \
  --keyring-backend test \
  --home "$HOME/.truerepublic"

./build/truerepublicd start \
  --home "$HOME/.truerepublic" \
  --minimum-gas-prices 0upnyx
```

Status prüfen:

```bash
curl -fsS http://127.0.0.1:26657/status | jq .result.sync_info
```

`SIGINT`/`Ctrl-C` fährt CometBFT und die Application-Datenbank geordnet herunter.
Ein erneuter Start mit demselben `--home` lädt Höhe und App-Hash persistent.

## Docker

```bash
docker compose build truerepublic-node
docker compose up -d truerepublic-node
docker compose logs -f truerepublic-node
```

Das Image läuft ohne Root-Rechte, initialisiert ein leeres benanntes Volume und
prüft den RPC-Status per Healthcheck. Details stehen in
[`node-operators/installation/docker-setup.md`](node-operators/installation/docker-setup.md).

## Maintained Web Client

```bash
cd client-web
npm ci
npm run lint
npm test -- --run
npm run dev
```

Die ZKP-Erzeugung im Web Client ist weiterhin ein Mock und nicht für echte
anonyme Abstimmungen freigegeben.

## CosmWasm-Verträge

```bash
cd contracts
cargo test --workspace
cargo build --release --target wasm32-unknown-unknown
```

Verträge erst nach einer bewusst konfigurierten Recovery-Testnet-Initialisierung
hochladen. Gebühren werden in `upnyx` angegeben.

## Bekannte Grenzen

- IBC-Staking und IBC-Upgrades verwenden weiterhin explizite Stubs.
- Multi-Node-Relayerbetrieb ist noch nicht recovery-verifiziert.
- Mainnet, reale Schlüssel und reale Gelder sind nicht freigegeben.
- Vollständige Grenzen: [`LIMITATIONS.md`](LIMITATIONS.md).

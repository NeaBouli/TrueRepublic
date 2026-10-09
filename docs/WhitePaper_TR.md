# TrueRepublic – White Paper

**Recovery edition v1 — 2026-10-10.** Versionierte Vision/Recovery-Dokumentation, kein Produktionsnachweis. Historische Modelle, Parameter und Roadmaps sind keine freigegebene Netzwerk-Spezifikation. Maßgeblich sind [Status](status.json), [Einschränkungen](LIMITATIONS.md) und [Rollout](ROLLOUT_ROADMAP.md); Produktions-ZKP, Setup-Provenienz und unabhängige Abnahme bleiben offen.

## 1. Einleitung & Vision
TrueRepublic (PNYX) ist eine Plattform für **direkte Demokratie**. Sie wurde entwickelt, um die Schwächen der repräsentativen Demokratie zu überwinden und den Bürgerinnen und Bürgern unmittelbare Mitbestimmung zu ermöglichen.

Kernidee: **digitale, zensurresistente, transparente und sichere Governance**, die auf Schwarmintelligenz setzt.

### Projektstatus und Lizenzierung

TrueRepublic ist ein gemeinschaftlich gesteuertes Open-Source-Projekt ohne
zentralen Firmeninhaber. Gepflegter Quellcode und gepflegte Dokumentation sind
Apache-2.0. Die Urheberrechte verbleiben bei den einzelnen Beitragenden; die
gemeinsame Zuordnung lautet „TrueRepublic contributors“. Markenmaterial,
Kunstwerke, historische PDFs, archivierte historische Nachweise und Material
Dritter bleiben ausgenommen, es sei denn, ein dateispezifischer Lizenzhinweis gilt
oder Herkunft und Erlaubnis sind dokumentiert. Diese
Lizenzgrundlage ändert weder den Rollout- noch den Produktionsstatus.

---

## 2. Problemstellung: Repräsentative Demokratie
- Macht konzentriert sich in wenigen Händen.
- Bürger geben ihre Stimme ab, verlieren danach aber Einfluss.
- Lobbyismus und Abhängigkeiten verzerren Entscheidungen.

**TrueRepublic** verfolgt direkte Beteiligung als Ziel. Die Zulässigkeit konkreter politischer oder rechtlicher Verfahren muss unabhängig und kontextbezogen geprüft werden; diese Vision garantiert keine Rechtskonformität.

---

## 3. TrueRepublic-Konzept
### 3.1 Proxy-Partei
Eine digitale Partei, die als „Vehikel“ dient, um direkte Demokratie in bestehende Systeme einzuführen.

### 3.2 Trustee-Modell
Geplante, nicht implementierte Vision: Stimmen könnten künftig einem weisungsgebundenen **Trustee** zugeordnet und zurückgenommen werden. Der aktuelle Recovery-Stand stellt keine Trustee-Delegation bereit.

### 3.3 Proof-of-Domain
Domänenbezogene Validator-/Stake-Regeln, keine Garantie rechtlich legitimer oder anonymer Abstimmungen. Gekaufte Coins sind nicht pauschal ausgeschlossen; siehe [Validator Guide](../wiki/operations/Validator-Guide.md). Produktions-ZKP und unabhängige Abnahme sind separate Voraussetzungen.

### 3.4 Schwarmintelligenz
Alle Vorschläge, Bewertungen und Abstimmungen erfolgen offen, überprüfbar und ohne zentrale Kontrolle.

---

## 4. Systemarchitektur
### 4.1 Blockchain (Cosmos SDK)
- Modul: `truedemocracy` (Domänen, Bewertungen und Stones; keine implementierte Trustee-Delegation)
- Modul: `dex` (dezentrale Börse für PNYX und IBC-Tokens)
- Modul: `treasury` (Kassenverwaltung, Gebühren, Belohnungen)

### 4.2 Smart Contracts (CosmWasm)
- Forschungsprototypen, keine belegten live betriebenen Produktionsverträge
- Kein Nachweis produktiver Governance oder sicherer Erweiterungen ohne Chain-Upgrade; siehe [Einschränkungen](LIMITATIONS.md)

### 4.3 Wallets
- **Web Client (React):** gepflegter Browser-Client unter `client-web`
- **Mobile Client (geplant):** Ein neuer iOS-/Android-Client benötigt vor einer
  Veröffentlichung sichere Schlüsselverwahrung und ein unabhängiges Wallet-/Krypto-Review;
  der frühere Prototyp wurde unter GH-102 stillgelegt.

---

## 5. Tokenomics
### 5.1 PNYX Token
- Utility & Governance Token
- Verwendung im Modell: Beteiligung, Gebühren und Domänen-Treasury; Trustee-Delegation bleibt geplant

### 5.2 Treasury
- Domänen-Einzahlungen und PayToPut finanzieren parametrisierte Beteiligungsbelohnungen; keine pauschale DEX-Gebühren-Weiterleitung an die Treasury wird behauptet
- Gemeinwohlprojekte/Entwicklungsfinanzierung sind Nutzungsideen, kein automatisch implementierter Auszahlungsworkflow

### 5.3 DEX
- AMM (Automated Market Maker)
- Pools: PNYX/ATOM, PNYX/IBC-Tokens
- Gebührenmodell: 0,3% Swap-Gebühr plus 1% PNYX-Burn, kein Maker/Taker-Treasury-Modell; siehe [Modulreferenz](developers/architecture/module-reference.md#dex-module)

---

## 6. Sicherheit & Compliance
- **Zensurresistenz:** Architekturziel, keine absolute Garantie; Mitgliedschaft, Berechtigungen, Client- und Netzwerkgrenzen bleiben relevant.
- **Transparenz:** Alle Transaktionen und Votes on-chain nachvollziehbar.
- **Datenschutz:** Personenbezogene Identitätsdaten sollen off-chain bleiben;
  auf der Chain werden, soweit möglich, nur Commitments, Nullifier und
  kryptografische Nachweise abgelegt. Pseudonyme Kennungen sind nicht
  automatisch anonym.
- **Prüfung:** Repository-interne Recovery- und Sicherheitsnachweise ersetzen
  keine unabhängige kryptografische, datenschutzrechtliche oder rechtliche
  Prüfung.

---

## 7. Roadmap & Offene Punkte
### Phase 1 (MVP)
- Blockchain (Cosmos SDK, v0.50.15)
- Basis-Module: Democracy, Treasury
- gepflegter Web-Client mit recovery-getestetem lokalem Wallet- und
  Transaktionspfad; die Prüfung eines Produktions-Signers bleibt offen

### Phase 2
- CosmWasm Contracts für Governance
- Sicherer Mobile-Client (Neuimplementierung; früherer Prototyp unter GH-102 stillgelegt)
- DEX (PNYX/ATOM, Slippage-Schutz)

### Phase 3
- Erweiterte DAO-Funktionalität
- Souveräne V4-Edge-Architektur mit TRChain als einziger Settlement-Chain,
  benutzerbetriebenen reduzierten Nodes, mobiler Verifikation, lokalen
  Gesetzes-/Diskussionsabläufen und abgeschotteten Domain-Apps gemäß der
  [V4-Architektur](SOVEREIGN_V4_EDGE_ARCHITECTURE.md). Sie enthält weder eine
  Minima-Abhängigkeit noch eine zweite Chain oder Bridge-Zusage.
- Optionale formale Abstimmungen mit systemischem Konsensieren,
  Ja/Nein/Enthaltung, Personenwahl und hybrider Ratifikation nach der
  [Wahlarchitektur](GOVERNANCE_BALLOT_ARCHITECTURE.md). GH-232 bleibt bis zur
  Stabilisierung des Rollouts und den erforderlichen ZKP-, Datenschutz- und
  Rechtsprozess-Prüfungen zurückgestellt.
- Treasury-Auszahlungslogik (Community Grants)
- Internationale Expansion

---

## 8. Zusammenfassung
**TrueRepublic / PNYX** bietet einen praktikablen Weg, direkte Demokratie in bestehende Systeme einzuführen – transparent, sicher, dezentral und skalierbar.

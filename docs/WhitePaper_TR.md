# TrueRepublic – White Paper

## 1. Einleitung & Vision
TrueRepublic (PNYX) ist eine Plattform für **direkte Demokratie**. Sie wurde entwickelt, um die Schwächen der repräsentativen Demokratie zu überwinden und den Bürgerinnen und Bürgern unmittelbare Mitbestimmung zu ermöglichen.

Kernidee: **digitale, zensurresistente, transparente und sichere Governance**, die auf Schwarmintelligenz setzt.

---

## 2. Problemstellung: Repräsentative Demokratie
- Macht konzentriert sich in wenigen Händen.
- Bürger geben ihre Stimme ab, verlieren danach aber Einfluss.
- Lobbyismus und Abhängigkeiten verzerren Entscheidungen.

**TrueRepublic** setzt hier an: Es dreht das System um, ohne Verfassung und Grundordnung brechen zu müssen.

---

## 3. TrueRepublic-Konzept
### 3.1 Proxy-Partei
Eine digitale Partei, die als „Vehikel“ dient, um direkte Demokratie in bestehende Systeme einzuführen.

### 3.2 Trustee-Modell
Bürger können ihre Stimme einem **Trustee** übertragen. Dieser handelt weisungsgebunden, die Stimme kann jederzeit zurückgenommen werden.

### 3.3 Proof-of-Domain
Ein neuartiger Mechanismus, der die Legitimität von Abstimmungen und Entscheidungen kryptographisch absichert.

### 3.4 Schwarmintelligenz
Alle Vorschläge, Bewertungen und Abstimmungen erfolgen offen, überprüfbar und ohne zentrale Kontrolle.

---

## 4. Systemarchitektur
### 4.1 Blockchain (Cosmos SDK)
- Modul: `truedemocracy` (Abstimmungen, Trustee-Logik)
- Modul: `dex` (dezentrale Börse für PNYX und IBC-Tokens)
- Modul: `treasury` (Kassenverwaltung, Gebühren, Belohnungen)

### 4.2 Smart Contracts (CosmWasm)
- Governance-Mechanismen (Proposals, Ratings, Tallying)
- Erweiterungen für neue Features ohne Chain-Upgrade

### 4.3 Wallets
- **Web Client (React):** gepflegter Browser-Client unter `client-web`
- **Mobile Client (geplant):** Ein neuer iOS-/Android-Client benötigt vor einer
  Veröffentlichung sichere Schlüsselverwahrung und ein unabhängiges Wallet-/Krypto-Review;
  der frühere Prototyp wurde unter GH-102 stillgelegt.

---

## 5. Tokenomics
### 5.1 PNYX Token
- Utility & Governance Token
- Verwendung: Stimmen abgeben, Trustees beauftragen, Gebühren zahlen, Treasury finanzieren

### 5.2 Treasury
- Einnahmen: Transaktionsgebühren, DEX-Gebühren
- Ausgaben: Gemeinwohlprojekte, Entwicklung, Community-Belohnungen

### 5.3 DEX
- AMM (Automated Market Maker)
- Pools: PNYX/ATOM, PNYX/IBC-Tokens
- Gebührenmodell: Maker/Taker + Treasury-Anteil

---

## 6. Sicherheit & Compliance
- **Zensurresistenz:** Keine zentrale Stelle kann Stimmen blockieren.
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
- Blockchain (Cosmos SDK, v0.50.13)
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

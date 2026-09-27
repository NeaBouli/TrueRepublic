verdict: changes
Substanz vollstaendig gegen Code verifiziert (f86f5a1): alle Hops Q1-Q5/
W1-W16, Modultabelle, Gaps G1-G4/G7 korrekt — u.a. escrow.go CacheContext/
write()-Ordnung, keeper.go ohne Member-Check (G3 regelkonform eskaliert),
CalcDomainCost=fee*2000 (G4), status.json 36/59, sovereignv4 unverdrahtet.
Regeln eingehalten: nur die 4 erlaubten Dateien, kein Code, kein Push;
check-consistency.sh + git diff --check reproduziert (exit 0); plantuml-
Skip regelkonform, Risiko offen deklariert. Drei kleine Maengel, fuer ein
evidenzbasiertes Dokument aber relevant:
1. MAP.md:20-22: README-Header belegt weder "domain-based governance" noch
   "capped PNYX" (Header: ZKP, CosmWasm, DEX; Cap nirgends in README.md).
2. MAP.md:131 "No file in docs/, wiki/ or README.md uses 'V3'" ist woertlich
   falsch (AGPLv3, yaml.v3, Versionsstrings); Kernaussage stimmt.
3. Modultabelle: Status "not traced" (und "built (external)") ausserhalb der
   sechs Brief-Werte — auf Enum mappen oder Abweichung deklarieren.
Kein Sicherheitsbefund im Diff selbst; Fixes sind je eine Zeile.

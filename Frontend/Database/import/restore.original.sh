#!/bin/bash

#needed in order to run this script
#sed -i 's/\r$//' restore.sh

/opt/mssql-tools/bin/sqlcmd -U 'sa' -P 'Start#123' -Q"RESTORE DATABASE PnyxAuthenticationDB FROM DISK='/var/opt/mssql/backup/PnyxAuthenticationDB.bak' WITH  FILE = 1, MOVE 'PnyxAuthenticationDB' TO '/var/opt/mssql/data/PnyxAuthenticationDB.mdf', MOVE 'PnyxAuthenticationDB_log' TO '/var/opt/mssql/data/PnyxAuthenticationDB_log.ldf', NOUNLOAD, REPLACE, STATS = 5;"
/opt/mssql-tools/bin/sqlcmd -U 'sa' -P 'Start#123' -Q"RESTORE DATABASE PnyxDB FROM DISK='/var/opt/mssql/backup/PnyxDb.bak' WITH  FILE = 1, MOVE 'PnyxDb' TO '/var/opt/mssql/data/PnyxDb.mdf', MOVE 'PnyxDb_log' TO '/var/opt/mssql/data/PnyxDb_log.ldf', NOUNLOAD, REPLACE, STATS = 5;"
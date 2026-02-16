#!/bin/bash
# TrueRepublic chain data backup script.
# Usage: ./scripts/backup.sh [backup-dir]
# Cron:  0 3 * * * /path/to/scripts/backup.sh

set -euo pipefail

CHAIN_HOME="${CHAIN_HOME:-$HOME/.truerepublic}"
BACKUP_DIR="${1:-$HOME/backup}"
DATE=$(date +%F)
BACKUP_FILE="truerepublic_${DATE}.tar.gz"
RETENTION_DAYS=30

mkdir -p "$BACKUP_DIR"

echo "[$(date)] Starting backup of $CHAIN_HOME..."

tar -czf "${BACKUP_DIR}/${BACKUP_FILE}" \
    -C "$(dirname "$CHAIN_HOME")" \
    "$(basename "$CHAIN_HOME")"

echo "[$(date)] Backup created: ${BACKUP_DIR}/${BACKUP_FILE}"

# Optional: upload to remote storage (uncomment and configure rclone)
# rclone copy "${BACKUP_DIR}/${BACKUP_FILE}" remote:TrueRepublicBackups

# Remove backups older than retention period
find "$BACKUP_DIR" -name "truerepublic_*.tar.gz" -mtime +${RETENTION_DAYS} -delete

echo "[$(date)] Cleanup complete. Backups older than ${RETENTION_DAYS} days removed."
echo "[$(date)] Backup finished successfully."

#!/usr/bin/env bash
# Rerunnable script to set up DB_VIEWER_APP_DATA table with seed preset data.
# Usage:
#   ORA_USER=LEARN1 ORA_PASS=Welcome ORA_HOST=localhost ORA_PORT=1521 ORA_SERVICE=XE ./db_viewer/sql_test/run_setup_app_data.sh
set -euo pipefail

ORA_USER="${ORA_USER:-LEARN1}"
ORA_PASS="${ORA_PASS:-Welcome}"
ORA_HOST="${ORA_HOST:-localhost}"
ORA_PORT="${ORA_PORT:-1521}"
ORA_SERVICE="${ORA_SERVICE:-XE}"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SQL_FILE="${SCRIPT_DIR}/setup_app_data.sql"

if [[ ! -f "$SQL_FILE" ]]; then
  echo "ERROR: SQL file not found: ${SQL_FILE}" >&2
  exit 2
fi

LOG_DIR="${SCRIPT_DIR}/logs"
mkdir -p "${LOG_DIR}"
TS="$(date '+%Y%m%d_%H%M%S')"
LOG_FILE="${LOG_DIR}/setup_app_data_${TS}.log"
TMP_SQL="${LOG_DIR}/_setup_app_data_tmp_${TS}.sql"

cat >"${TMP_SQL}" <<EOSQL
SET ECHO ON FEEDBACK ON SERVEROUTPUT ON LINESIZE 200 PAGESIZE 100 TRIMSPOOL ON VERIFY OFF
SPOOL ${LOG_FILE}
PROMPT === Running setup_app_data.sql ===
@${SQL_FILE}
SPOOL OFF
EXIT
EOSQL

CONN_STRING="${ORA_USER}/${ORA_PASS}@${ORA_HOST}:${ORA_PORT}/${ORA_SERVICE}"

if command -v sql >/dev/null 2>&1; then
  echo "[INFO] Using SQLcl (sql)"
  sql -S "$CONN_STRING" @"${TMP_SQL}"
elif command -v sqlplus >/dev/null 2>&1; then
  echo "[INFO] Using SQL*Plus (sqlplus)"
  sqlplus -s "$CONN_STRING" @"${TMP_SQL}"
else
  echo "ERROR: Neither SQLcl (sql) nor SQL*Plus (sqlplus) found in PATH." >&2
  echo "Generated SQL driver at: ${TMP_SQL}" >&2
  echo "You can run it manually with your Oracle client." >&2
  exit 3
fi

echo "[OK] DB_VIEWER_APP_DATA setup completed. Log: ${LOG_FILE}"

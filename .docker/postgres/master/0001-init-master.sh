#!/usr/bin/env bash
# PostgreSQL 18 primary setup: replication + SCRAM + roles
set -Eeuo pipefail

: "${PGDATA:?Defina PGDATA}"
: "${POSTGRES_USER:?Defina POSTGRES_USER}"
: "${REPLICATION_USER:?Defina REPLICATION_USER}"
: "${REPLICATION_PASSWORD:?Defina REPLICATION_PASSWORD}"

# Role de app opcional (NOLOGIN) para grants futuros
APP_ROLE="${APP_ROLE:-app_role}"

CONF="$PGDATA/postgresql.conf"
HBA="$PGDATA/pg_hba.conf"

# Helper para tornar idempotente: substitui se existir, senão acrescenta
set_conf() {
	local key="$1" val="$2"
	if grep -Eq "^[[:space:]]*${key}[[:space:]]*=" "$CONF"; then
		sed -ri "s|^[[:space:]]*${key}[[:space:]]*=.*|${key} = ${val}|" "$CONF"
	else
		printf "%s = %s\n" "$key" "$val" >>"$CONF"
	fi
}

echo "[*] Ajustando postgresql.conf…"
set_conf "listen_addresses" "'*'"
# Replicação física (streaming)
set_conf "wal_level" "replica"        # v18: padrão é 'replica' (setado explicitamente)
set_conf "max_wal_senders" "10"       # número de conexões de standbys
set_conf "max_replication_slots" "10" # útil para slots físicos/lógicos
set_conf "wal_keep_size" "'1GB'"      # retenção mínima de WAL para standbys atrasados
# Segurança de senhas moderna (SCRAM)
set_conf "password_encryption" "'scram-sha-256'"

echo "[*] Ajustando pg_hba.conf…"
# Linhas para replicação (IPv4 e IPv6) usando SCRAM
if ! grep -qE "^host[[:space:]]+replication[[:space:]]+${REPLICATION_USER}[[:space:]]+0\\.0\\.0\\.0/0[[:space:]]+scram-sha-256" "$HBA"; then
	printf "host replication %s 0.0.0.0/0 scram-sha-256\n" "$REPLICATION_USER" >>"$HBA"
fi
if ! grep -qE "^host[[:space:]]+replication[[:space:]]+${REPLICATION_USER}[[:space:]]+::/0[[:space:]]+scram-sha-256" "$HBA"; then
	printf "host replication %s ::/0 scram-sha-256\n" "$REPLICATION_USER" >>"$HBA"
fi

echo "[*] Recarregando configuração…"
psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d postgres -c "SELECT pg_reload_conf();"

echo "[*] Criando usuário de replicação e role de aplicação…"
psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d postgres <<SQL
DO \$\$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = '${REPLICATION_USER}') THEN
    -- Usuário de replicação precisa de LOGIN + REPLICATION
    CREATE ROLE ${REPLICATION_USER} LOGIN REPLICATION PASSWORD '${REPLICATION_PASSWORD}';
  ELSE
    ALTER ROLE ${REPLICATION_USER} LOGIN REPLICATION PASSWORD '${REPLICATION_PASSWORD}';
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = '${APP_ROLE}') THEN
    -- Role genérica sem login para grants (ajuste ao seu uso)
    CREATE ROLE ${APP_ROLE} NOLOGIN;
  END IF;
END
\$\$;
SQL

create_db_if_missing() {
        local dbname="$1"

        if [ -z "$dbname" ]; then
                return
        fi

        if ! psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d postgres -tc "SELECT 1 FROM pg_database WHERE datname = '$dbname'" | grep -q 1; then
                echo "[*] Criando database '$dbname'…"
                psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d postgres -c "CREATE DATABASE \"$dbname\" OWNER $POSTGRES_USER;"
        fi
}

create_db_if_missing "$POSTGRES_DB"
create_db_if_missing "$GATEWAY_POSTGRES_DB"
create_db_if_missing "$TENANT_POSTGRES_DB"
create_db_if_missing "$GUARDRAIL_POSTGRES_DB"
create_db_if_missing "$OBSERVABILITY_POSTGRES_DB"

echo "[ok] Primário configurado para streaming replication (PostgreSQL 18)."

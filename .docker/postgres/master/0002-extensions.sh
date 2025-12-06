#!/usr/bin/env bash
# Apply required extensions to every configured database.
set -euo pipefail

SQL_FILE="/docker-entrypoint-initdb.d/0002-extensions.sql"

apply_extensions() {
        local dbname="$1"

        if [[ -z "$dbname" ]]; then
                return
        fi

        echo "[*] Applying extensions on database '$dbname'..."
        psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$dbname" -f "$SQL_FILE"
}

declare -A visited=()
for db in "${POSTGRES_DB:-}" "${GATEWAY_POSTGRES_DB:-}" "${TENANT_POSTGRES_DB:-}" \
        "${GUARDRAIL_POSTGRES_DB:-}" "${OBSERVABILITY_POSTGRES_DB:-}"; do
        if [[ -n "$db" && -z "${visited[$db]:-}" ]]; then
                visited[$db]=1
                apply_extensions "$db"
        fi
done


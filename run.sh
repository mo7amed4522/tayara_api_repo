#!/usr/bin/env bash
# run.sh — local dev bootstrap for the Tayara API.
#
# 1. Reads config.yaml (mysql/redis/mongo section) for ports, credentials, db names.
# 2. Asks which infra containers to create/start via docker compose
#    (mysql / redis / mongo / all / none — choose none if you already run them locally).
# 3. Asks whether to start the API server with `go run main.go`.
#
# Usage: ./run.sh            (interactive)
#        ./run.sh all        (non-interactive: all infra + no server prompt is skipped)
set -euo pipefail
cd "$(dirname "$0")"

CONFIG_FILE="config.yaml"
COMPOSE_FILE="docker-compose.yml"

# ---------------------------------------------------------------- helpers ----
# yq-free yaml getter: reads `section:` then `key:` (2-space indented, comments stripped).
yaml_get() { # yaml_get <section> <key>
    awk -v section="$1" -v key="$2" '
        $0 ~ "^"section":"            { insec=1; next }
        /^[a-zA-Z]/                   { insec=0 }
        insec && $0 ~ "^[[:space:]]+"key":" {
            sub("^[^:]*:[[:space:]]*", "")
            sub(/[[:space:]]+#.*$/, "")        # strip trailing comment
            gsub(/^"|"$/, "")                  # strip surrounding quotes
            print; exit
        }
    ' "$CONFIG_FILE"
}

confirm() { # confirm <prompt> <default y|n>
    local prompt="$1" default="${2:-y}" reply
    if [[ "$default" == "y" ]]; then prompt="$prompt [Y/n] "; else prompt="$prompt [y/N] "; fi
    read -r -p "$prompt" reply
    reply="${reply:-$default}"
    [[ "$reply" =~ ^[Yy]$ ]]
}

die() { echo "error: $*" >&2; exit 1; }

# ------------------------------------------------------------ read config ----
[[ -f "$CONFIG_FILE" ]]   || die "$CONFIG_FILE not found"
[[ -f "$COMPOSE_FILE" ]]  || die "$COMPOSE_FILE not found"

MYSQL_PORT=$(yaml_get mysql port);            MYSQL_PORT=${MYSQL_PORT:-3306}
MYSQL_DB=$(yaml_get mysql db-name);           MYSQL_DB=${MYSQL_DB:-tayara}
MYSQL_USER=$(yaml_get mysql username);        MYSQL_USER=${MYSQL_USER:-root}
MYSQL_PASS=$(yaml_get mysql password);        MYSQL_PASS=${MYSQL_PASS:-tayara_2026}

REDIS_ADDR=$(yaml_get redis addr)             # host:port
REDIS_PORT=${REDIS_ADDR##*:}; REDIS_PORT=${REDIS_PORT:-6379}
REDIS_PASS=$(yaml_get redis password)

MONGO_PORT=$(yaml_get mongo hosts)            # hosts has its own key — fallback below
MONGO_PORT=$(awk '/^mongo:/{f=1} f&&/port:/{gsub(/[ "]/,"",$2); print $2; exit}' "$CONFIG_FILE")
MONGO_PORT=${MONGO_PORT:-27017}
MONGO_DB=$(yaml_get mongo database);          MONGO_DB=${MONGO_DB:-tayara}
MONGO_USER=$(yaml_get mongo username);        MONGO_USER=${MONGO_USER:-admin}
MONGO_PASS=$(yaml_get mongo password);        MONGO_PASS=${MONGO_PASS:-tayara_2026}

# map config.yaml values onto the env vars docker-compose.yml understands
export MYSQL_PORT MYSQL_DATABASE="$MYSQL_DB" MYSQL_USER MYSQL_PASSWORD="$MYSQL_PASS" \
       MYSQL_ROOT_PASSWORD="$MYSQL_PASS" \
       REDIS_PORT MONGO_PORT MONGO_DATABASE="$MONGO_DB" MONGO_USER="$MONGO_USER" MONGO_PASSWORD="$MONGO_PASS"

echo "config.yaml → mysql:$MYSQL_PORT db=$MYSQL_DB | redis:$REDIS_PORT | mongo:$MONGO_PORT db=$MONGO_DB"

# ------------------------------------------------------- pick infra services --
CHOICE="${1:-}"
if [[ -z "$CHOICE" ]]; then
    echo
    echo "Which infrastructure containers do you want to create/start?"
    echo "  1) mysql   2) redis   3) mongo   4) all   5) none (already installed locally)"
    read -r -p "Choice [4]: " CHOICE
    CHOICE=${CHOICE:-4}
fi

SERVICES=()
case "$CHOICE" in
    1|mysql)        SERVICES=(mysql) ;;
    2|redis)        SERVICES=(redis) ;;
    3|mongo)        SERVICES=(mongo) ;;
    4|all|"")       SERVICES=(mysql redis mongo) ;;
    5|none|skip)    SERVICES=() ;;
    *)              die "unknown choice: $CHOICE (use mysql|redis|mongo|all|none)" ;;
esac

if [[ ${#SERVICES[@]} -gt 0 ]]; then
    command -v docker >/dev/null || die "docker not found — install Docker Desktop or pick 'none'"
    echo "→ docker compose up -d ${SERVICES[*]}"
    docker compose -f "$COMPOSE_FILE" up -d "${SERVICES[@]}"

    echo "→ waiting for containers to become healthy…"
    for svc in "${SERVICES[@]}"; do
        for _ in $(seq 1 30); do
            status=$(docker inspect -f '{{.State.Health.Status}}' "tayara-$svc" 2>/dev/null || echo "missing")
            [[ "$status" == "healthy" ]] && break
            sleep 2
        done
        echo "   $svc: $status"
        [[ "$status" == "healthy" ]] || die "$svc did not become healthy — check: docker compose logs $svc"
    done
else
    echo "→ skipping containers (using your locally installed mysql/redis/mongo)"
fi

# ------------------------------------------------------------- run server? ---
echo
if confirm "Start the API server now (go run main.go)?" "y"; then
    exec go run main.go
fi

echo "Done. Start later with: go run main.go"

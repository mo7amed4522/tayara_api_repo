#!/usr/bin/env bash
# run.sh — local dev bootstrap for the Tayara API, no Docker.
#
# Starts, in order, whatever is not already listening:
#   1. MySQL   (local tarball install, mysqld on :3306)
#   2. Redis   (local install, redis-server on :6379)
#   3. Kafka   (local install, broker on :9092) + the four support topics
# then runs the API with `go run main.go` (HTTP on :8888).
#
# Every service is skipped when its port is already open, so re-running is safe.
# Paths can be overridden: MYSQL_HOME=… MYSQL_DATA=… REDIS_HOME=… KAFKA_HOME=… ./run.sh
#
# Usage: ./run.sh
set -euo pipefail
cd "$(dirname "$0")"

CONFIG_FILE="config.yaml"
LOG_DIR="log"
mkdir -p "$LOG_DIR"

# ---------------------------------------------------------------- helpers ----
# yq-free yaml getter: reads `section:` then `key:` (2-space indented, comments stripped).
yaml_get() { # yaml_get <section> <key>
    awk -v section="$1" -v key="$2" '
        $0 ~ "^"section":"            { insec=1; next }
        /^[a-zA-Z]/                   { insec=0 }
        insec && $0 ~ "^[[:space:]]+"key":" {
            sub("^[^:]*:[[:space:]]*", "")
            sub(/[[:space:]]+#.*$/, "")
            gsub(/^"|"$/, "")
            print; exit
        }
    ' "$CONFIG_FILE"
}

# Resolve a yaml value that may be written as ${VAR}, ${VAR:-default} or ${VAR-default}.
env_value() {
    local raw="$1" name
    if [[ "$raw" =~ ^\$\{([A-Za-z_][A-Za-z0-9_]*):-(.*)\}$ ]]; then
        name="${BASH_REMATCH[1]}"; printf '%s' "${!name:-${BASH_REMATCH[2]}}"
    elif [[ "$raw" =~ ^\$\{([A-Za-z_][A-Za-z0-9_]*)-(.*)\}$ ]]; then
        name="${BASH_REMATCH[1]}"; printf '%s' "${!name-${BASH_REMATCH[2]}}"
    elif [[ "$raw" =~ ^\$\{([A-Za-z_][A-Za-z0-9_]*)\}$ ]]; then
        printf '%s' "${!BASH_REMATCH[1]:-}"
    else
        printf '%s' "$raw"
    fi
}

port_open() { nc -z 127.0.0.1 "$1" >/dev/null 2>&1; }

# Start a long-running process detached: ignores Ctrl+C / hangup so stopping the API
# (the last line of this script) leaves MySQL, Redis and Kafka running.
spawn() { # spawn <logfile> <command...>
    local logfile="$1"; shift
    ( trap '' HUP INT; exec "$@" ) >>"$logfile" 2>&1 &
    disown || true
}

wait_port() { # wait_port <port> <label> [tries]
    local port="$1" label="$2" tries="${3:-60}" i
    for ((i = 0; i < tries; i++)); do
        if port_open "$port"; then echo "   $label ready on :$port"; return 0; fi
        sleep 1
    done
    return 1
}

die() { echo "error: $*" >&2; exit 1; }

# ------------------------------------------------------------ read config ----
[[ -f "$CONFIG_FILE" ]] || die "$CONFIG_FILE not found"

MYSQL_PORT=$(env_value "$(yaml_get mysql port)"); MYSQL_PORT=${MYSQL_PORT:-3306}
REDIS_ADDR=$(env_value "$(yaml_get redis addr)")
REDIS_PORT=${REDIS_ADDR##*:}; REDIS_PORT=${REDIS_PORT:-6379}

# Paths from this machine (override with env vars).
MYSQL_HOME="${MYSQL_HOME:-/Volumes/Untitled/mysql/mysql-26.7.0-macos15-arm64}"
MYSQL_DATA="${MYSQL_DATA:-/Volumes/Untitled/mysql/data}"
MYSQL_SOCKET="${MYSQL_SOCKET:-/tmp/mysql.sock}"
REDIS_HOME="${REDIS_HOME:-/Volumes/Untitled/redis}"
REDIS_CONF="${REDIS_CONF:-$REDIS_HOME/redis-7.4.6/redis.conf}"
KAFKA_HOME="${KAFKA_HOME:-}"

echo "local bootstrap → mysql:$MYSQL_PORT | redis:$REDIS_PORT | kafka:9092 | api:8888"

# ----------------------------------------------------------------- MySQL -----
if port_open "$MYSQL_PORT"; then
    echo "→ MySQL already running on :$MYSQL_PORT"
else
    [[ -x "$MYSQL_HOME/bin/mysqld" ]] || die "mysqld not found at $MYSQL_HOME/bin/mysqld (set MYSQL_HOME)"
    echo "→ starting MySQL from $MYSQL_HOME (log: $LOG_DIR/mysqld.log)"
    spawn "$LOG_DIR/mysqld.log" "$MYSQL_HOME/bin/mysqld" \
        --basedir="$MYSQL_HOME" \
        --datadir="$MYSQL_DATA" \
        --port="$MYSQL_PORT" \
        --socket="$MYSQL_SOCKET" \
        --console
    wait_port "$MYSQL_PORT" "MySQL" 60 || die "MySQL did not open :$MYSQL_PORT — check $LOG_DIR/mysqld.log"
fi

# ----------------------------------------------------------------- Redis -----
if port_open "$REDIS_PORT"; then
    echo "→ Redis already running on :$REDIS_PORT"
else
    [[ -x "$REDIS_HOME/bin/redis-server" ]] || die "redis-server not found at $REDIS_HOME/bin/redis-server (set REDIS_HOME)"
    [[ -f "$REDIS_CONF" ]] || die "redis config not found at $REDIS_CONF (set REDIS_CONF)"
    echo "→ starting Redis from $REDIS_HOME (log: $LOG_DIR/redis.log)"
    ROOT_DIR=$(pwd -P)
    ( cd "$REDIS_HOME" && spawn "$ROOT_DIR/$LOG_DIR/redis.log" ./bin/redis-server "$REDIS_CONF" )
    wait_port "$REDIS_PORT" "Redis" 45 || die "Redis did not open :$REDIS_PORT — check $LOG_DIR/redis.log"
fi

# ----------------------------------------------------------------- Kafka -----
# Find a local Kafka install: explicit KAFKA_HOME, then the usual spots.
if [[ -z "$KAFKA_HOME" ]]; then
    for candidate in /Volumes/Untitled/kafka* /Volumes/Untitled/kafka_[0-9]* "$HOME"/kafka* "$HOME"/Downloads/kafka* \
                     /opt/homebrew/opt/kafka/libexec /usr/local/opt/kafka/libexec /opt/kafka; do
        if [[ -x "${candidate%/}/bin/kafka-server-start.sh" ]]; then KAFKA_HOME="${candidate%/}"; break; fi
    done
fi

if port_open 9092; then
    echo "→ Kafka already running on :9092"
elif [[ -n "$KAFKA_HOME" && -x "$KAFKA_HOME/bin/kafka-server-start.sh" ]]; then
    KAFKA_CONFIG="$KAFKA_HOME/config/kraft/server.properties"
    [[ -f "$KAFKA_CONFIG" ]] || die "KRaft config not found: $KAFKA_CONFIG"
    KAFKA_LOG_DIRS=$(awk -F= '/^[[:space:]]*log\.dirs[[:space:]]*=/{gsub(/[[:space:]]/, "", $2); print $2}' "$KAFKA_CONFIG")
    KAFKA_LOG_DIRS=${KAFKA_LOG_DIRS%%,*}
    if [[ ! -f "${KAFKA_LOG_DIRS}/meta.properties" ]]; then
        echo "→ formatting Kafka KRaft storage in $KAFKA_LOG_DIRS"
        CLUSTER_ID=$("$KAFKA_HOME/bin/kafka-storage.sh" random-uuid)
        "$KAFKA_HOME/bin/kafka-storage.sh" format -t "$CLUSTER_ID" -c "$KAFKA_CONFIG" >/dev/null
    fi
    echo "→ starting Kafka from $KAFKA_HOME (log: $LOG_DIR/kafka.log)"
    spawn "$LOG_DIR/kafka.log" "$KAFKA_HOME/bin/kafka-server-start.sh" "$KAFKA_CONFIG"
    wait_port 9092 "Kafka" 90 || die "Kafka did not open :9092 — check $LOG_DIR/kafka.log"
else
    # Nothing on :9092 and no install to launch: run this session without the bus so the
    # fanout consumers do not retry-dial a dead port forever. Chat still works (DB + local hub).
    export KAFKA_ENABLED=false
    echo "warning: no local Kafka found (looked for bin/kafka-server-start.sh in KAFKA_HOME, /Volumes/Untitled/kafka*, \$HOME/kafka*, Homebrew, /opt/kafka)."
    echo "         Starting the API with the bus disabled for this run; messages persist and fan out locally."
    echo "         Point the script at your install with: KAFKA_HOME=/path/to/kafka ./run.sh"
fi

# Create the support topics (auto-create is disabled in config.yaml).
if port_open 9092; then
    if [[ -n "$KAFKA_HOME" && -x "$KAFKA_HOME/bin/kafka-topics.sh" ]]; then
        echo "→ ensuring support topics exist"
        for topic in tayara.support.chat.messages.v1 tayara.support.chat.events.v1 \
                     tayara.support.chat.messages.dlq.v1 tayara.support.presence.v1; do
            "$KAFKA_HOME/bin/kafka-topics.sh" --bootstrap-server localhost:9092 \
                --create --if-not-exists --topic "$topic" --partitions 3 --replication-factor 1 \
                >/dev/null 2>&1 || echo "   could not create $topic"
        done
    else
        echo "warning: Kafka is up but kafka-topics.sh was not found; create the four tayara.support.* topics manually."
    fi
fi

# ------------------------------------------------------------------ API -----
echo "→ starting API (go run main.go) on :8888"
exec go run main.go

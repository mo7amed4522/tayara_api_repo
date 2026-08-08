#!/usr/bin/env bash
#
# Tayara API — defensive security attack/scan script
# ------------------------------------------------------------------
# Probes a running Tayara API instance for the confirmed vulnerabilities
# found during the security review and reports VULNERABLE / FIXED per check.
# Run AFTER the fixes are applied to confirm they hold; run BEFORE to
# reproduce the issues.
#
# Usage:
#   ./security/attack.sh                      # defaults to http://localhost:8888
#   ./security/attack.sh http://192.168.1.10:8888
#
# Each check is a standalone curl probe + offline grep where a live call is
# not applicable. The script never modifies data — it only sends harmless
# probe requests (registration attempts that should fail, throwaway logins,
# forged-token probes, malformed uploads).
# ------------------------------------------------------------------
set -uo pipefail

BASE_URL="${1:-http://localhost:8888}"
PASS=0
FAIL=0
SKIP=0

up() {
  curl -s -o /dev/null -w '%{http_code}' --max-time 3 "$BASE_URL/api/v1/base/categories" 2>/dev/null | grep -qE '200|4[0-9][0-9]'
}

# --- output helpers -------------------------------------------------
result() { # result "label" [detail]
  case "$1" in
    FIXED)      echo "  [ PASS ] $2 ${3:+($3)}"; PASS=$((PASS+1));;
    VULNERABLE) echo "  [ FAIL ] $2 ${3:+($3)}"; FAIL=$((FAIL+1));;
    SKIP)       echo "  [ SKIP ] $2 ${3:+($3)}"; SKIP=$((SKIP+1));;
  esac
}

# --- crypto helpers (HS256 JWT forging with a candidate key) --------
b64url() { openssl base64 -A | tr '+/' '-_' | tr -d '='; }
b64url_raw() { printf '%s' "$1" | b64url; }
hmac_sign() { printf '%s' "$1" | openssl dgst -sha256 -hmac "$2" -binary | b64url; }

# forge_jwt <key>  ->  prints an admin S-token signed with <key>
forge_jwt() {
  local key="$1"
  local header='{"alg":"HS256","typ":"JWT"}'
  # exp = 2100-01-01, so the token is never expired during the scan
  local payload='{"ID":1,"Username":"admin","NickName":"admin","AuthorityId":1,"BufferTime":0,"exp":4102444800,"iat":1720000000,"iss":"tayara-api"}'
  local h="$(b64url_raw "$header")" p="$(b64url_raw "$payload")"
  local sig="$(hmac_sign "$h.$p" "$key")"
  printf '%s.%s.%s' "$h" "$p" "$sig"
}

section() { echo ""; echo "== $1 =="; }

echo "Target: $BASE_URL"
if up; then
  echo "Server reachable — running live probes."
else
  echo "WARNING: server not reachable. Live probes will SKIP; static checks still run."
fi

# =====================================================================
# CWE-284 Improper Access Control — open platform-admin registration
# =====================================================================
section "1. Broken Access Control: open /system/register (CWE-284)"
if up; then
  RAND="scan$RANDOM$RANDOM"
  code=$(curl -s -o /tmp/tayara_reg.json -w '%{http_code}' --max-time 5 \
    -X POST "$BASE_URL/api/v1/system/register" \
    -H 'Content-Type: application/json' \
    -d "{\"userName\":\"$RAND\",\"password\":\"ScanPass_123\",\"nickName\":\"scan\",\"email\":\"$RAND@tayara.test\",\"phoneCountryCode\":\"+971\",\"phone\":\"0500000000\"}")
  body=$(cat /tmp/tayara_reg.json)
  if echo "$body" | grep -qiE '"(user|admin|id)":|registration.*success|"code"\s*:\s*0'; then
    result VULNERABLE "open admin registration" "HTTP $code — a new platform admin was created"
  else
    result FIXED "open admin registration" "HTTP $code — registration rejected"
  fi
else
  result SKIP "open admin registration"
fi

# =====================================================================
# CWE-327 Cryptographic Failures — weak/guessable JWT signing key
# =====================================================================
section "2. Cryptographic Failure: forgeable JWT signing key (CWE-327)"
if up; then
  # Try known-weak/default keys. If ANY produces a 200 on an admin endpoint,
  # the deployment is using a guessable key -> tokens are forgeable.
  hit=""
  for key in tayar tayara tayara-prod secret 123456 ""; do
    tok="$(forge_jwt "$key")"
    code=$(curl -s -o /dev/null -w '%{http_code}' --max-time 5 "$BASE_URL/api/v1/system/jobs" -H "S-token: $tok")
    if [ "$code" = "200" ]; then hit="$key"; break; fi
  done
  if [ -n "$hit" ]; then
    result VULNERABLE "weak JWT key forgery" "token signed with \"$hit\" accepted as admin"
  else
    result FIXED "weak JWT key forgery" "no guessable key forged a valid admin token"
  fi
else
  result SKIP "weak JWT key forgery"
fi

# =====================================================================
# CWE-307 Improper Restriction of Excessive Authentication Attempts
# =====================================================================
section "3. Brute-force protection on login (CWE-307)"
if up; then
  throttled=0; total=0
  for i in $(seq 1 15); do
    code=$(curl -s -o /dev/null -w '%{http_code}' --max-time 3 \
      -X POST "$BASE_URL/api/v1/user/login" \
      -H 'Content-Type: application/json' \
      -d '{"email":"nobody@tayara.test","password":"wrong"}')
    total=$((total+1))
    if [ "$code" = "429" ] || [ "$code" = "403" ]; then throttled=$((throttled+1)); fi
  done
  if [ "$throttled" -ge 3 ]; then
    result FIXED "login rate limit" "server throttled $throttled/$total rapid attempts (HTTP 429)"
  else
    result VULNERABLE "login rate limit" "$throttled/$total throttled — brute-force not blocked"
  fi
else
  result SKIP "login rate limit"
fi

# =====================================================================
# CWE-204 Improper Output Neutralization for Logs / user enumeration
# =====================================================================
section "4. Account enumeration via login error messages (CWE-204)"
if up; then
  known=$(curl -s --max-time 5 -X POST "$BASE_URL/api/v1/system/login" \
    -H 'Content-Type: application/json' -d '{"email":"admin@tayara.test","password":"wrongpass_zz"}')
  unknown=$(curl -s --max-time 5 -X POST "$BASE_URL/api/v1/system/login" \
    -H 'Content-Type: application/json' -d '{"email":"no-such-user-$RANDOM@tayara.test","password":"wrongpass_zz"}')
  # Normalize the messages: strip any dynamic tokens so we compare content only.
  known_n=$(echo "$known"   | sed -E 's/[0-9a-f]{6,}//g; s/"[0-9]+"/"N"/g')
  unk_n=$(echo "$unknown"   | sed -E 's/[0-9a-f]{6,}//g; s/"[0-9]+"/"N"/g')
  if [ "$known_n" = "$unk_n" ]; then
    result FIXED "login enumeration" "identical response for known vs unknown account"
  else
    result VULNERABLE "login enumeration" "responses differ (e.g. \"record not found\" vs \"password error\")"
  fi
else
  result SKIP "login enumeration"
fi

# =====================================================================
# CWE-327 — JWT algorithm-confusion (alg=none / non-HS256)
# =====================================================================
section "5. JWT algorithm confusion (CWE-327)"
if up; then
  # alg=none token (no signature)
  h=$(b64url_raw '{"alg":"none","typ":"JWT"}')
  p=$(b64url_raw '{"ID":1,"Username":"admin","exp":4102444800}')
  none_tok="$h.$p."
  code=$(curl -s -o /dev/null -w '%{http_code}' --max-time 5 "$BASE_URL/api/v1/system/jobs" -H "S-token: $none_tok")
  if [ "$code" = "200" ]; then
    result VULNERABLE "alg=none accepted" "HTTP 200"
  else
    result FIXED "alg=none rejected" "HTTP $code"
  fi
else
  result SKIP "alg=none probe"
fi

# =====================================================================
# Static (offline) checks — no server required
# =====================================================================
section "6. Static checks (source/config) — no server required"

# 6a. upload allowlist present?
if grep -q 'allowedUploadExts' utils/upload/local.go 2>/dev/null; then
  result FIXED "upload extension allowlist"
else
  result VULNERABLE "upload extension allowlist" "no allowlist in utils/upload/local.go"
fi

# 6b. upload path-traversal guard present?
if grep -q 'invalid upload path\|strings.ContainsAny(customPath' utils/upload/local.go 2>/dev/null; then
  result FIXED "upload path-traversal guard"
else
  result VULNERABLE "upload path-traversal guard"
fi

# 6c. admin-registration fail-closed present?
if grep -q 'System admin registration is disabled' api/v1/system/sys_admin_profile.go 2>/dev/null; then
  result FIXED "admin registration fail-closed"
else
  result VULNERABLE "admin registration fail-closed"
fi

# 6d. auth rate-limit wired?
if grep -q 'AuthRateLimit' initialize/router_biz.go middleware/bruteforce_limit.go 2>/dev/null; then
  result FIXED "auth rate-limit middleware"
else
  result VULNERABLE "auth rate-limit middleware"
fi

# 6e. JWT algorithm pinned to HS256?
if grep -q 'SigningMethodHS256' utils/jwt.go utils/bx_jwt.go 2>/dev/null; then
  result FIXED "JWT HS256 algorithm pinned"
else
  result VULNERABLE "JWT HS256 algorithm pinned"
fi

# 6f. strong-key boot guard present?
if grep -q 'validateJWTSigningKey' main.go 2>/dev/null; then
  result FIXED "JWT signing-key strength guard"
else
  result VULNERABLE "JWT signing-key strength guard"
fi

# 6g. MSSQL autocode SQLi sanitizer present?
if grep -q 'sanitizeIdent' service/system/sys_auto_code_mssql.go 2>/dev/null; then
  result FIXED "MSSQL autocode SQLi sanitizer"
else
  result VULNERABLE "MSSQL autocode SQLi sanitizer"
fi

# 6h. firebase service-account key untracked?
if git ls-files --error-unmatch firebase/tayara-service-account.json >/dev/null 2>&1; then
  result VULNERABLE "firebase private key untracked" "still tracked in git"
elif [ -f firebase/tayara-service-account.json ]; then
  result FIXED "firebase private key untracked" "ignored on disk (rotate the key!)"
else
  result FIXED "firebase private key untracked" "not present"
fi

# 6i. weak JWT key still committed as a default?
if grep -rqE 'signing-key:\s*(tayar|tayara|tayara-prod)\b' config.yaml 2>/dev/null; then
  result VULNERABLE "weak JWT key in config.yaml" "default signing-key is short/guessable"
else
  result FIXED "weak JWT key in config.yaml"
fi

# =====================================================================
echo ""
echo "==============================="
echo " Summary: $PASS passed, $FAIL vulnerable, $SKIP skipped"
echo "==============================="
[ "$FAIL" -eq 0 ]

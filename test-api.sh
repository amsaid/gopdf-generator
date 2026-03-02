#!/bin/bash
# ================================================
#  GoPDF Generator — API Test Suite
# ================================================

BASE_URL="${BASE_URL:-http://localhost:8080}"
OUTPUT_DIR="${OUTPUT_DIR:-./test-output}"

# ── Colors & styles ──────────────────────────────
BOLD='\033[1m'
DIM='\033[2m'
RESET='\033[0m'
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'

# ── Symbols ──────────────────────────────────────
PASS="${GREEN}✔${RESET}"
FAIL="${RED}✘${RESET}"
ARROW="${CYAN}❯${RESET}"

# ── State ────────────────────────────────────────
PASS_COUNT=0
FAIL_COUNT=0
SKIP_COUNT=0

# ── Helpers ──────────────────────────────────────
section() {
  echo ""
  echo -e "${BLUE}${BOLD}┌─ $1 ${RESET}"
}

log_pass() { echo -e "  ${PASS} ${GREEN}$1${RESET}"; ((PASS_COUNT++)); }
log_fail() { echo -e "  ${FAIL} ${RED}$1${RESET}";   ((FAIL_COUNT++)); }
log_skip() { echo -e "  ${YELLOW}⊘${RESET} ${DIM}$1${RESET}"; ((SKIP_COUNT++)); }
log_info() { echo -e "  ${ARROW} ${CYAN}$1${RESET}"; }
log_cmd()  { echo -e "  ${DIM}\$ $1${RESET}"; }
log_kv()   { printf "  ${DIM}%-18s${RESET} %s\n" "$1" "$2"; }

print_json() {
  if command -v jq &>/dev/null; then
    echo "$1" | jq . 2>/dev/null | sed 's/^/    /' || echo "    $1"
  else
    echo "$1" | sed 's/^/    /'
  fi
}

# test_request <label> <method> <path> [body] [binary]
test_request() {
  local label="$1" method="$2" path="$3" body="$4" binary="$5"
  local url="${BASE_URL}${path}"

  echo ""
  echo -e "  ${ARROW} ${BOLD}${label}${RESET}"
  log_cmd "curl -s -X ${method} \"${url}\""

  local start_ms end_ms elapsed http_code response

  start_ms=$(date +%s%3N 2>/dev/null || echo 0)

  if [[ "$method" == "POST" && -n "$body" ]]; then
    if [[ "$binary" == "true" ]]; then
      local outfile="${OUTPUT_DIR}/api-test.pdf"
      http_code=$(curl -s -w "%{http_code}" -X POST "$url" \
        -H "Content-Type: application/json" -d "$body" -o "$outfile")
      response="[binary → ${outfile}]"
    else
      local raw
      raw=$(curl -s -w "\n__CODE__%{http_code}" -X POST "$url" \
        -H "Content-Type: application/json" -d "$body")
      http_code=$(echo "$raw" | grep '__CODE__' | sed 's/.*__CODE__//')
      response=$(echo "$raw" | grep -v '__CODE__')
    fi
  else
    local raw
    raw=$(curl -s -w "\n__CODE__%{http_code}" "$url")
    http_code=$(echo "$raw" | grep '__CODE__' | sed 's/.*__CODE__//')
    response=$(echo "$raw" | grep -v '__CODE__')
  fi

  end_ms=$(date +%s%3N 2>/dev/null || echo 0)
  elapsed=$(( end_ms - start_ms ))

  if [[ -z "$http_code" || "$http_code" == "000" ]]; then
    log_fail "Connection refused — is the server running at ${BASE_URL}?"
    return 1
  fi

  if [[ "$http_code" -ge 200 && "$http_code" -lt 300 ]]; then
    log_pass "HTTP ${http_code} · ${elapsed}ms"
    if [[ "$binary" == "true" ]]; then
      log_info "Saved to ${OUTPUT_DIR}/api-test.pdf"
    elif [[ -n "$response" ]]; then
      print_json "$response"
    fi
    return 0
  else
    log_fail "HTTP ${http_code} · ${elapsed}ms"
    [[ -n "$response" ]] && print_json "$response"
    return 1
  fi
}

preflight() {
  if ! command -v curl &>/dev/null; then
    log_fail "curl is not installed"; exit 1
  else
    log_pass "curl $(curl --version | head -1 | awk '{print $2}')"
  fi

  if ! command -v jq &>/dev/null; then
    log_skip "jq not found — JSON output will be raw"
  else
    log_pass "jq $(jq --version)"
  fi

  mkdir -p "$OUTPUT_DIR"
  log_pass "Output dir: ${OUTPUT_DIR}"
}

summary() {
  echo ""
  echo -e "  ${DIM}$(printf '═%.0s' {1..48})${RESET}"
  echo ""
  echo -e "  ${BOLD}Results${RESET}"
  log_kv "Passed:"  "${GREEN}${PASS_COUNT}${RESET}"
  log_kv "Failed:"  "${RED}${FAIL_COUNT}${RESET}"
  log_kv "Skipped:" "${YELLOW}${SKIP_COUNT}${RESET}"
  echo ""
  if [[ $FAIL_COUNT -eq 0 ]]; then
    echo -e "  ${GREEN}${BOLD}All tests passed!${RESET}"
  else
    echo -e "  ${RED}${BOLD}${FAIL_COUNT} test(s) failed.${RESET}"
  fi
  echo ""
}

# ── Main ─────────────────────────────────────────
clear
echo ""
echo -e "${BOLD}${WHITE}  ╔══════════════════════════════════════════╗"
echo -e "  ║     GoPDF Generator — API Test Suite    ║"
echo -e "  ╚══════════════════════════════════════════╝${RESET}"
echo ""
log_kv "Base URL:" "${CYAN}${BASE_URL}${RESET}"
log_kv "Output:"   "${CYAN}${OUTPUT_DIR}${RESET}"

section "Preflight"
preflight

section "Test 1 · Health"
test_request "Health endpoint" GET "/health"

section "Test 2 · Fonts"
test_request "Available fonts" GET "/api/v1/fonts"

section "Test 3 · Template Validation"
test_request "Validate template schema" POST "/api/v1/templates/validate" \
  '{"page_size":"A4","elements":[{"type":"text","text":"Test"}]}'

section "Test 4 · PDF Generation"
test_request "Generate PDF from template" POST "/api/v1/generate/template" \
  '{"page_size":"A4","margin":{"top":50,"bottom":50,"left":50,"right":50},"elements":[{"type":"text","text":"API Test","font":{"size":20}}]}' \
  true

summary

[[ $FAIL_COUNT -gt 0 ]] && exit 1 || exit 0

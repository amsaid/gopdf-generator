#!/bin/bash
# ================================================
#  GoPDF Generator — Build & Test Suite
# ================================================

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
START_TIME=$(date +%s)

# ── Helpers ──────────────────────────────────────
section() {
  echo -e "\n${BLUE}${BOLD}┌─ $1${RESET}"
}

log_pass() { echo -e "  ${PASS} ${GREEN}$1${RESET}"; ((PASS_COUNT++)); }
log_fail() { echo -e "  ${FAIL} ${RED}$1${RESET}";   ((FAIL_COUNT++)); }
log_skip() { echo -e "  ${YELLOW}⊘${RESET} ${DIM}$1${RESET}"; ((SKIP_COUNT++)); }
log_info() { echo -e "  ${ARROW} ${CYAN}$1${RESET}"; }
log_cmd()  { echo -e "  ${DIM}\$ $1${RESET}"; }
# Fix: Use %b to interpret color codes in $2
log_kv()   { printf "  ${DIM}%-20s${RESET} %b\n" "$1" "$2"; }

run_step() {
  local label="$1"; shift
  echo -e "\n  ${ARROW} ${BOLD}${label}${RESET}"
  log_cmd "$*"

  # Use pipefail to ensure if the command fails, the whole pipeline returns non-zero
  set -o pipefail
  if "$@" 2>&1 | sed 's/^/    /'; then
    log_pass "${label} completed"
    return 0
  else
    log_fail "${label} failed"
    return 1
  fi
}

summary() {
  local end_time=$(date +%s)
  local elapsed=$(( end_time - START_TIME ))
  echo -e "\n  ${DIM}────────────────────────────────────────────────${RESET}"
  echo -e "  ${BOLD}Results${RESET}"
  log_kv "Passed:"   "${GREEN}${PASS_COUNT}${RESET}"
  log_kv "Failed:"   "${RED}${FAIL_COUNT}${RESET}"
  log_kv "Skipped:"  "${YELLOW}${SKIP_COUNT}${RESET}"
  log_kv "Duration:" "${WHITE}${elapsed}s${RESET}"
  echo ""
  if [[ $FAIL_COUNT -eq 0 ]]; then
    echo -e "  ${GREEN}${BOLD}All steps passed!${RESET}"
  else
    echo -e "  ${RED}${BOLD}${FAIL_COUNT} step(s) failed.${RESET}"
  fi
  echo ""
}

# ── Main ─────────────────────────────────────────
tput clear || clear

echo -e "${BOLD}${WHITE}  ╔══════════════════════════════════════════╗"
echo -e "  ║    GoPDF Generator — Build & Test Suite  ║"
echo -e "  ╚══════════════════════════════════════════╝${RESET}"

# ── Preflight ────────────────────────────────────
section "Preflight"

if ! command -v go &>/dev/null; then
  log_fail "Go is not installed"
  echo -e "  ${YELLOW}Install Go 1.24+:${RESET} https://golang.org/dl/"
  exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
log_pass "Go ${GO_VERSION}"

G_PATH=$(go env GOPATH 2>/dev/null || echo "Not Set")
G_OS=$(go env GOOS 2>/dev/null || echo "Unknown")
G_ARCH=$(go env GOARCH 2>/dev/null || echo "Unknown")

log_kv "GOPATH:" "${CYAN}${G_PATH}${RESET}"
log_kv "Platform:" "${CYAN}${G_OS}/${G_ARCH}${RESET}"

# ── Setup ─────────────────────────────────────────
section "Setup"

echo -e "  ${ARROW} ${BOLD}Checking permissions & directories${RESET}"
if mkdir -p bin fonts downloads test-output 2>/dev/null; then
  log_pass "Directories bin/ fonts/ downloads/ test-output/ ready"
else
  log_fail "Could not create directories. Check folder permissions."
  exit 1
fi

if [[ -f "go.mod" ]]; then
    run_step "Download dependencies" go mod download
else
    log_skip "go.mod not found (skipping dependencies)"
fi

# ── Tests ─────────────────────────────────────────
section "Tests"

# Standardize: run_step handles the log_pass/fail internally
run_step "Run test suite" go test -v ./pkg/parser/... ./pkg/rtl/... ./pkg/generator/... || true

# ── Build ─────────────────────────────────────────
section "Build"

if run_step "Build server binary" go build -o bin/gopdf-server cmd/server/main.go; then
  # Removed 'local' because we aren't in a function here
  build_size=$(du -sh bin/gopdf-server 2>/dev/null | awk '{print $1}')
  log_info "Binary size: ${build_size}  →  bin/gopdf-server"
fi

# ── Examples ───────────────────────────
section "Examples"

if [[ -f "examples/usage_example.go" ]]; then
  run_step "Run usage examples" go run examples/usage_example.go || true
else
  log_skip "examples/usage_example.go not found"
fi

# ── Summary ───────────────────────────────────────
summary

# ── Next steps ────────────────────────────────────
if [[ $FAIL_COUNT -eq 0 ]]; then
    echo -e "  ${BOLD}Next steps${RESET}"
    echo -e "  ${DIM}Start the server:${RESET} ${CYAN}./bin/gopdf-server${RESET}\n"
fi

[[ $FAIL_COUNT -gt 0 ]] && exit 1 || exit 0

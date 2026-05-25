#!/usr/bin/env bash

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BOLD='\033[1m'
RESET='\033[0m'

MIN_COVERAGE=${MIN_COVERAGE:-85}
FAILED=0

# Strict coverage enforced only on these packages
PACKAGES=(
  "./internal/common/shared/..."
)

# Dynamically discover all domain packages
while IFS= read -r dir; do
  PACKAGES+=("./${dir}/...")
done < <(find internal -type d -name "domain" | sort)

echo -e "${BOLD}Packages under coverage enforcement (≥${MIN_COVERAGE}%):${RESET}"
for pkg in "${PACKAGES[@]}"; do
  echo -e "  ${YELLOW}→${RESET} $pkg"
done
echo ""

rm -f coverage_*.out

for pattern in "${PACKAGES[@]}"; do
  while IFS= read -r pkg; do
    echo -e "${BOLD}Testing:${RESET} $pkg"

    COVER_FILE="coverage_$(echo "$pkg" | tr '/' '_').out"

    set +e
    go test \
      -race \
      -timeout=2m \
      -coverprofile="$COVER_FILE" \
      -covermode=atomic \
      "$pkg"
    TEST_EXIT=$?
    set -e

    if [[ $TEST_EXIT -ne 0 ]]; then
      echo -e "${RED}${BOLD}FAIL${RESET}${RED} Tests failed for $pkg${RESET}\n"
      FAILED=1
      continue
    fi

    COVERAGE=$(go tool cover -func="$COVER_FILE" \
      | grep -E '^total:' \
      | awk '{print $3}' \
      | sed 's/%//')

    if [[ -z "$COVERAGE" ]]; then
      echo -e "${YELLOW}SKIP${RESET} $pkg => no statements\n"
      continue
    fi

    if awk "BEGIN {exit !($COVERAGE >= $MIN_COVERAGE)}"; then
      echo -e "${GREEN}${BOLD}PASS${RESET} $pkg => ${GREEN}${COVERAGE}%${RESET}\n"
    else
      echo -e "${RED}${BOLD}FAIL${RESET} $pkg => ${RED}${COVERAGE}%${RESET} ${YELLOW}(minimum: ${MIN_COVERAGE}%)${RESET}\n"
      FAILED=1
    fi

  done < <(go list "$pattern" 2>/dev/null)
done

rm -f coverage_*.out

if [[ $FAILED -eq 1 ]]; then
  echo -e "${RED}${BOLD}Coverage check failed.${RESET}"
  exit 1
fi

echo -e "${GREEN}${BOLD}All coverage checks passed.${RESET}"
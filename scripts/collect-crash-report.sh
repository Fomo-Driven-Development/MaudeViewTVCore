#!/bin/sh
#
# collect-crash-report.sh - Bundle controller/chromium diagnostics for crash triage
#
# Output:
#   logs/crash-report-<timestamp>.tar.gz
#

set -eu

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
LOG_DIR="${PROJECT_ROOT}/logs"
CRASH_DIR="${LOG_DIR}/chromium-crash-dumps"
STAMP="$(date +%Y%m%d-%H%M%S)"
WORK_DIR="${LOG_DIR}/crash-report-${STAMP}"
ARCHIVE="${LOG_DIR}/crash-report-${STAMP}.tar.gz"

mkdir -p "${WORK_DIR}"

# Snapshot runtime state.
{
    echo "timestamp=${STAMP}"
    echo ""
    echo "==== process snapshot ===="
    ps -ef | grep -E "bin/controller|chromium|chrome" | grep -v grep || true
    echo ""
    echo "==== local endpoints ===="
    curl -s -i --max-time 2 http://127.0.0.1:8188/health || true
    echo ""
    curl -s -i --max-time 2 http://127.0.0.1:9220/json/version || true
} > "${WORK_DIR}/snapshot.txt" 2>&1

# Copy key logs if present.
for f in \
    "${LOG_DIR}/controller.log" \
    "${LOG_DIR}/chromium-debug.log" \
    "${LOG_DIR}/chromium.pid"; do
    if [ -f "${f}" ]; then
        cp "${f}" "${WORK_DIR}/"
    fi
done

# Always include tails for quick scan.
if [ -f "${LOG_DIR}/controller.log" ]; then
    tail -n 400 "${LOG_DIR}/controller.log" > "${WORK_DIR}/controller.tail.log" || true
fi
if [ -f "${LOG_DIR}/chromium-debug.log" ]; then
    tail -n 600 "${LOG_DIR}/chromium-debug.log" > "${WORK_DIR}/chromium-debug.tail.log" || true
fi

# Copy crash dumps if available.
if [ -d "${CRASH_DIR}" ]; then
    mkdir -p "${WORK_DIR}/chromium-crash-dumps"
    cp -a "${CRASH_DIR}/." "${WORK_DIR}/chromium-crash-dumps/" || true
fi

tar -czf "${ARCHIVE}" -C "${LOG_DIR}" "crash-report-${STAMP}"
rm -rf "${WORK_DIR}"

echo "Created crash report: ${ARCHIVE}"

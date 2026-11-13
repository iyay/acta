#!/usr/bin/env bash
set -euo pipefail
exec bash "$(dirname "$0")/../land-gate-output-approved/scaffold.sh" no-approval

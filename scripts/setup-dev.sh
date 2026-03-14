#!/usr/bin/env bash
set -euo pipefail

pnpm install
make -C apps/api test

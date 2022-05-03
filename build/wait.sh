#!/bin/bash

set -e

__workdir="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
__rootdir=$(dirname "${__workdir}")

cd "${__rootdir}"

until docker-compose exec -T postgres psql --user=postgres -c "SELECT 1;" >/dev/null 2>&1; do
  echo "Waiting for database connection..."
  sleep 1
done
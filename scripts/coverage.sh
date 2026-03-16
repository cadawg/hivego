#!/usr/bin/env bash
set -e

echo "=== Unit tests ==="
go test ./... -coverprofile=cover_unit.out
go tool cover -func=cover_unit.out | tail -1

echo ""
echo "=== Integration tests ==="
go test -tags integration ./... -coverprofile=cover_integration.out
go tool cover -func=cover_integration.out | tail -1

rm -f cover_unit.out cover_integration.out
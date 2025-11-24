#!/usr/bin/env bash
set -e

# One-click start (Linux/macOS): build and start all services, then show status.
docker-compose up -d --build
docker-compose ps

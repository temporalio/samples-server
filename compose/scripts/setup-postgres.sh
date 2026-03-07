#!/bin/sh
# @@@SNIPSTART compose-postgres-setup
set -eu

# Validate required environment variables
: "${POSTGRES_SEEDS:?ERROR: POSTGRES_SEEDS environment variable is required}"
: "${POSTGRES_USER:?ERROR: POSTGRES_USER environment variable is required}"

echo 'Starting PostgreSQL schema setup...'
echo 'Waiting for PostgreSQL port to be available...'
nc -z -w 10 ${POSTGRES_SEEDS} ${DB_PORT:-5432}
echo 'PostgreSQL port is available'

# Create and setup temporal database
temporal-sql-tool --plugin postgres12 --ep ${POSTGRES_SEEDS} -u ${POSTGRES_USER} -p ${DB_PORT:-5432} --db temporal create
temporal-sql-tool --plugin postgres12 --ep ${POSTGRES_SEEDS} -u ${POSTGRES_USER} -p ${DB_PORT:-5432} --db temporal setup-schema -v 0.0
temporal-sql-tool --plugin postgres12 --ep ${POSTGRES_SEEDS} -u ${POSTGRES_USER} -p ${DB_PORT:-5432} --db temporal update-schema -d /etc/temporal/schema/postgresql/v12/temporal/versioned

# Create and setup visibility database
temporal-sql-tool --plugin postgres12 --ep ${POSTGRES_SEEDS} -u ${POSTGRES_USER} -p ${DB_PORT:-5432} --db temporal_visibility create
temporal-sql-tool --plugin postgres12 --ep ${POSTGRES_SEEDS} -u ${POSTGRES_USER} -p ${DB_PORT:-5432} --db temporal_visibility setup-schema -v 0.0
temporal-sql-tool --plugin postgres12 --ep ${POSTGRES_SEEDS} -u ${POSTGRES_USER} -p ${DB_PORT:-5432} --db temporal_visibility update-schema -d /etc/temporal/schema/postgresql/v12/visibility/versioned

echo 'PostgreSQL schema setup complete'
# @@@SNIPEND
#!/bin/sh
set -eu

echo 'Starting MySQL schema setup...'
echo 'Waiting for MySQL port to be available...'
nc -z -w 10 mysql 3306
echo 'MySQL port is available'

# Create and setup temporal database
temporal-sql-tool --plugin mysql8 --ep mysql -u root -p 3306 --db temporal create
temporal-sql-tool --plugin mysql8 --ep mysql -u root -p 3306 --db temporal setup-schema -v 0.0
temporal-sql-tool --plugin mysql8 --ep mysql -u root -p 3306 --db temporal update-schema -d /etc/temporal/schema/mysql/v8/temporal/versioned

# Create and setup visibility database
temporal-sql-tool --plugin mysql8 --ep mysql -u root -p 3306 --db temporal_visibility create
temporal-sql-tool --plugin mysql8 --ep mysql -u root -p 3306 --db temporal_visibility setup-schema -v 0.0
temporal-sql-tool --plugin mysql8 --ep mysql -u root -p 3306 --db temporal_visibility update-schema -d /etc/temporal/schema/mysql/v8/visibility/versioned

echo 'MySQL schema setup complete'

#!/bin/sh
set -eu

# Validate required environment variables
: "${ES_SCHEME:?ERROR: ES_SCHEME environment variable is required}"
: "${ES_HOST:?ERROR: ES_HOST environment variable is required}"
: "${ES_PORT:?ERROR: ES_PORT environment variable is required}"
: "${ES_VISIBILITY_INDEX:?ERROR: ES_VISIBILITY_INDEX environment variable is required}"
: "${ES_VERSION:?ERROR: ES_VERSION environment variable is required}"
: "${ES_USER:?ERROR: ES_USER environment variable is required}"
: "${ES_PWD:?ERROR: ES_PWD environment variable is required}"
: "${POSTGRES_USER:?ERROR: POSTGRES_USER environment variable is required}"

echo 'Starting PostgreSQL and Elasticsearch (TLS) schema setup...'
echo 'Waiting for PostgreSQL port to be available...'
nc -z -w 10 postgresql ${POSTGRES_DEFAULT_PORT:-5432}
echo 'PostgreSQL port is available'

# Create and setup temporal database with TLS
temporal-sql-tool --plugin postgres12 --ep postgresql -u ${POSTGRES_USER} -p ${POSTGRES_DEFAULT_PORT:-5432} --db temporal --tls --tls-ca-file /usr/local/share/ca-certificates/ca.crt create
temporal-sql-tool --plugin postgres12 --ep postgresql -u ${POSTGRES_USER} -p ${POSTGRES_DEFAULT_PORT:-5432} --db temporal --tls --tls-ca-file /usr/local/share/ca-certificates/ca.crt setup-schema -v 0.0
temporal-sql-tool --plugin postgres12 --ep postgresql -u ${POSTGRES_USER} -p ${POSTGRES_DEFAULT_PORT:-5432} --db temporal --tls --tls-ca-file /usr/local/share/ca-certificates/ca.crt update-schema -d /etc/temporal/schema/postgresql/v12/temporal/versioned

# Setup Elasticsearch index with TLS
# temporal-elasticsearch-tool is available in v1.30+ server releases
if [ -x /usr/local/bin/temporal-elasticsearch-tool ]; then
  echo 'Using temporal-elasticsearch-tool for Elasticsearch setup'
  temporal-elasticsearch-tool --ep "$ES_SCHEME://$ES_HOST:$ES_PORT" --user "$ES_USER" --password "$ES_PWD" setup-schema
  temporal-elasticsearch-tool --ep "$ES_SCHEME://$ES_HOST:$ES_PORT" --user "$ES_USER" --password "$ES_PWD" create-index --index $ES_VISIBILITY_INDEX
else
  echo 'Using curl for Elasticsearch setup'
  echo 'WARNING: curl will be removed from admin-tools in v1.30.'
  echo 'Waiting for Elasticsearch to be ready...'
  max_attempts=30
  attempt=0
  until curl -s -f -k -u "$ES_USER:$ES_PWD" "$ES_SCHEME://$ES_HOST:$ES_PORT/_cluster/health?wait_for_status=yellow&timeout=1s"; do
    attempt=$((attempt + 1))
    if [ $attempt -ge $max_attempts ]; then
      echo "ERROR: Elasticsearch did not become ready after $max_attempts attempts"
      echo "Last error from curl:"
      curl -k -u "$ES_USER:$ES_PWD" "$ES_SCHEME://$ES_HOST:$ES_PORT/_cluster/health?wait_for_status=yellow&timeout=1s" 2>&1 || true
      exit 1
    fi
    echo "Elasticsearch not ready yet, waiting... (attempt $attempt/$max_attempts)"
    sleep 2
  done
  echo ''
  echo 'Elasticsearch is ready'
  echo 'Creating index template...'
  curl -X PUT --fail -k -u "$ES_USER:$ES_PWD" "$ES_SCHEME://$ES_HOST:$ES_PORT/_template/temporal_visibility_v1_template" -H 'Content-Type: application/json' --data-binary "@/etc/temporal/schema/elasticsearch/visibility/index_template_$ES_VERSION.json"
  echo ''
  echo 'Creating index...'
  curl -k -u "$ES_USER:$ES_PWD" --head --fail "$ES_SCHEME://$ES_HOST:$ES_PORT/$ES_VISIBILITY_INDEX" 2>/dev/null || curl -k -u "$ES_USER:$ES_PWD" -X PUT --fail "$ES_SCHEME://$ES_HOST:$ES_PORT/$ES_VISIBILITY_INDEX"
  echo ''
fi

echo 'PostgreSQL and Elasticsearch (TLS) setup complete'

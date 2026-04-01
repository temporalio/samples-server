#!/bin/sh
set -eu
NAMESPACE=${DEFAULT_NAMESPACE:-default}
TEMPORAL_ADDRESS=${TEMPORAL_ADDRESS:-temporal:7233}

TEMPORAL_HOST=$(echo $TEMPORAL_ADDRESS | cut -d: -f1)
TEMPORAL_PORT=$(echo $TEMPORAL_ADDRESS | cut -d: -f2)

echo "Waiting for Temporal server port to be available..."
nc -z -w 10 $TEMPORAL_HOST $TEMPORAL_PORT
echo 'Temporal server port is available'

# Resolve hostname to IPv4 to work around some very weird issue
TEMPORAL_IP=$(getent ahosts $TEMPORAL_HOST | grep STREAM | head -1 | awk '{print $1}')
RESOLVED_ADDRESS="${TEMPORAL_IP}:${TEMPORAL_PORT}"
echo "Resolved $TEMPORAL_ADDRESS to $RESOLVED_ADDRESS"

echo 'Waiting for Temporal server to be healthy...'
max_attempts=3
attempt=0
until temporal operator cluster health --address $RESOLVED_ADDRESS; do
  attempt=$((attempt + 1))
  if [ $attempt -ge $max_attempts ]; then
    echo "Server did not become healthy after $max_attempts attempts"
    exit 1
  fi
  echo "Server not ready yet, waiting... (attempt $attempt/$max_attempts)"
  sleep 5
done

echo "Server is healthy, creating namespace '$NAMESPACE'..."
temporal operator namespace describe -n $NAMESPACE --address $RESOLVED_ADDRESS || temporal operator namespace create -n $NAMESPACE --address $RESOLVED_ADDRESS
echo "Namespace '$NAMESPACE' created"

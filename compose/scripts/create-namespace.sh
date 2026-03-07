#!/bin/sh
set -eu

NAMESPACE=${DEFAULT_NAMESPACE:-default}
TEMPORAL_ADDRESS=${TEMPORAL_ADDRESS:-temporal:7233}

echo "Waiting for Temporal server port to be available..."
nc -z -w 10 $(echo $TEMPORAL_ADDRESS | cut -d: -f1) $(echo $TEMPORAL_ADDRESS | cut -d: -f2)
echo 'Temporal server port is available'

echo 'Waiting for Temporal server to be healthy...'
max_attempts=3
attempt=0

until temporal operator cluster health --address $TEMPORAL_ADDRESS; do
  attempt=$((attempt + 1))
  if [ $attempt -ge $max_attempts ]; then
    echo "Server did not become healthy after $max_attempts attempts"
    exit 1
  fi
  echo "Server not ready yet, waiting... (attempt $attempt/$max_attempts)"
  sleep 5
done

echo "Server is healthy, creating namespace '$NAMESPACE'..."
temporal operator namespace describe -n $NAMESPACE --address $TEMPORAL_ADDRESS || temporal operator namespace create -n $NAMESPACE --address $TEMPORAL_ADDRESS
echo "Namespace '$NAMESPACE' created"

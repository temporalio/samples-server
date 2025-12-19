#!/bin/bash

set -e

# Configuration
TEMPORAL_ADDRESS="${TEMPORAL_ADDRESS:-temporal:7233}"
NAMESPACE="${NAMESPACE:-default}"
MAX_WAIT_TIME="${MAX_WAIT_TIME:-300}"

echo "=== Temporal Validation Script ==="
echo "Temporal address: $TEMPORAL_ADDRESS"
echo "Namespace: $NAMESPACE"
echo "Max wait time: ${MAX_WAIT_TIME}s"
echo

# Wait a bit for Temporal to be ready
echo "Waiting for Temporal server to be accessible..."
sleep 15

# Check cluster health
echo "=== Checking cluster health ==="
if OUTPUT=$(temporal operator cluster health --address "$TEMPORAL_ADDRESS" 2>&1); then
    echo "✓ Cluster health check passed"
    echo "$OUTPUT"
else
    echo "✗ Cluster health check failed" >&2
    echo "Error output:" >&2
    echo "$OUTPUT" >&2
    exit 1
fi

# Describe namespace to verify it exists and is accessible
echo
echo "=== Verifying namespace '$NAMESPACE' ==="
if OUTPUT=$(temporal operator namespace describe --namespace "$NAMESPACE" --address "$TEMPORAL_ADDRESS" 2>&1); then
    echo "✓ Namespace '$NAMESPACE' is accessible"
else
    echo "✗ Namespace '$NAMESPACE' not found or not accessible" >&2
    echo "Error output:" >&2
    echo "$OUTPUT" >&2
    exit 1
fi

# Try to start a simple workflow to validate full functionality
echo
echo "=== Testing workflow execution ==="
WORKFLOW_ID="validation-test-$(date +%s)"
TASK_QUEUE="validation-queue"

# Start a workflow (this will fail gracefully if no workers, but proves the system is working)
if OUTPUT=$(temporal workflow start \
    --workflow-id "$WORKFLOW_ID" \
    --type "NonExistentWorkflow" \
    --task-queue "$TASK_QUEUE" \
    --namespace "$NAMESPACE" \
    --address "$TEMPORAL_ADDRESS" \
    --execution-timeout 10s 2>&1); then
    echo "✓ Successfully initiated workflow (proves Temporal is functional)"
    echo "$OUTPUT"

    # Clean up - terminate the workflow since it won't complete
    echo "Terminating test workflow..."
    temporal workflow terminate \
        --workflow-id "$WORKFLOW_ID" \
        --namespace "$NAMESPACE" \
        --address "$TEMPORAL_ADDRESS" \
        --reason "Validation complete" || true
else
    # Workflow start failed, but check if workflow was actually created (means server is working)
    echo "Workflow start command returned non-zero, checking if workflow exists..."
    echo "Workflow start output:"
    echo "$OUTPUT"

    if DESCRIBE_OUTPUT=$(temporal workflow describe \
        --workflow-id "$WORKFLOW_ID" \
        --namespace "$NAMESPACE" \
        --address "$TEMPORAL_ADDRESS" 2>&1); then
        echo "✓ Workflow was created and server is functional"
        echo "Terminating test workflow..."
        temporal workflow terminate \
            --workflow-id "$WORKFLOW_ID" \
            --namespace "$NAMESPACE" \
            --address "$TEMPORAL_ADDRESS" \
            --reason "Validation complete" || true
    else
        echo "✗ Failed to create workflow" >&2
        echo "Workflow describe output:" >&2
        echo "$DESCRIBE_OUTPUT" >&2
        exit 1
    fi
fi

echo
echo "=== All validation checks passed! ==="
exit 0

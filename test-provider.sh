#!/bin/bash
set -e

echo "Testing provider-rabbitmq deployment"
echo "====================================="

if ! kubectl cluster-info &>/dev/null; then
    echo "No Kubernetes cluster available. Please ensure kubectl is configured."
    exit 1
fi

if ! kubectl get deployment crossplane -n crossplane-system &>/dev/null; then
    echo "Crossplane not found. Install it first:"
    echo "  helm repo add crossplane-stable https://charts.crossplane.io/stable"
    echo "  helm install crossplane crossplane-stable/crossplane --namespace crossplane-system --create-namespace"
    exit 1
fi

echo "Kubernetes cluster and Crossplane found"

echo "Checking CRDs..."
CRDS=(
    "vhosts.rabbitmq.crossplane.io"
    "exchanges.rabbitmq.crossplane.io"
    "queues.rabbitmq.crossplane.io"
    "bindings.rabbitmq.crossplane.io"
    "users.rabbitmq.crossplane.io"
    "permissions.rabbitmq.crossplane.io"
    "providerconfigs.rabbitmq.crossplane.io"
    "providerconfigusages.rabbitmq.crossplane.io"
)

all_ok=true
for crd in "${CRDS[@]}"; do
    if kubectl get crd "$crd" &>/dev/null; then
        echo "  OK  $crd"
    else
        echo "  MISSING  $crd"
        all_ok=false
    fi
done

if [ "$all_ok" = false ]; then
    echo "Some CRDs are missing. Apply the package CRDs first."
    exit 1
fi

echo ""
echo "Checking provider pod..."
POD_NAME=$(kubectl get pods -n crossplane-system -l app=provider-rabbitmq -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")
if [ -n "$POD_NAME" ]; then
    echo "Provider pod: $POD_NAME"
    kubectl logs -n crossplane-system "$POD_NAME" --tail=20
else
    echo "Provider pod not found"
fi

echo ""
echo "Next steps:"
echo "  1. Apply credentials: kubectl apply -f examples/provider/secret.yaml.tmpl"
echo "  2. Apply provider config: kubectl apply -f examples/provider/providerconfig.yaml"
echo "  3. Apply sample resources: kubectl apply -f examples/sample-resources.yaml"

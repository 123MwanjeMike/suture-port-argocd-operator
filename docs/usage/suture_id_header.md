# Setting SUTURE_ID Environment Variable

This example demonstrates how to configure the ArgoCD Operator to add a custom `Suture_ID` header to all HTTP requests made to the Kubernetes API server.

## Overview

When the `SUTURE_ID` environment variable is set, the operator will automatically add a `Suture_ID` header to all HTTP requests. This is useful for:
- Request tracking and correlation
- Distributed tracing
- Debugging and troubleshooting
- Audit logging

## Configuration

### Option 1: Setting via Deployment

If you're deploying the operator manually, you can set the environment variable in the operator's Deployment:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: argocd-operator-controller-manager
  namespace: argocd-operator-system
spec:
  template:
    spec:
      containers:
      - name: manager
        image: quay.io/argoprojlabs/argocd-operator:latest
        env:
        - name: SUTURE_ID
          value: "my-tracking-id-12345"
        # ... other environment variables
```

### Option 2: Setting via OLM Subscription

If you're using Operator Lifecycle Manager (OLM), you can set the environment variable in the Subscription:

```yaml
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: argocd-operator
  namespace: argocd-operator-system
spec:
  channel: alpha
  name: argocd-operator
  source: argocd-catalog
  sourceNamespace: olm
  config:
    env:
    - name: SUTURE_ID
      value: "my-tracking-id-12345"
```

### Option 3: Dynamic Value from ConfigMap or Secret

You can also source the value from a ConfigMap or Secret:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: argocd-operator-controller-manager
  namespace: argocd-operator-system
spec:
  template:
    spec:
      containers:
      - name: manager
        image: quay.io/argoprojlabs/argocd-operator:latest
        env:
        - name: SUTURE_ID
          valueFrom:
            configMapKeyRef:
              name: operator-config
              key: suture-id
        # OR from a secret
        # - name: SUTURE_ID
        #   valueFrom:
        #     secretKeyRef:
        #       name: operator-secrets
        #       key: suture-id
```

## Verification

To verify that the header is being added, you can:

1. Check the operator logs for any HTTP request traces
2. Use a network interceptor or proxy to inspect outgoing requests
3. Enable Kubernetes API server audit logging to see incoming request headers

## Example Use Cases

### Distributed Tracing
```yaml
env:
- name: SUTURE_ID
  value: "trace-id-$(POD_NAME)-$(POD_NAMESPACE)"
```

### Environment-Based Tracking
```yaml
env:
- name: SUTURE_ID
  value: "prod-cluster-01"
```

### Dynamic ID Generation
```yaml
env:
- name: SUTURE_ID
  valueFrom:
    fieldRef:
      fieldPath: metadata.uid
```

## Notes

- If `SUTURE_ID` is not set or is empty, no header will be added
- The header is added to all HTTP requests made by the operator to the Kubernetes API
- This does not affect requests made by ArgoCD components themselves, only the operator

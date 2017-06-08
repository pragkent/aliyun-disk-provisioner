# aliyun-disk-provisioner

[![Build Status](https://travis-ci.org/pragkent/aliyun-disk-provisioner.svg?branch=master)](https://travis-ci.org/pragkent/aliyun-disk-provisioner)

Aliyun Disk Kubernetes Dynamic Provisioner

## Usage

1. Deploy aliyundisk-provisioner to kubernetes cluster:
```bash
kubectl apply -f deploy/rbac.yaml
kubectl apply -f deploy/secret.yaml # Need to fillin your own aliyun access key
kubectl apply -f deploy/deployment.yaml
kubectl apply -f deploy/storage_class.yaml
```

2. Use it to dynamic provsioning persistent volumes.

```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: my-disk
spec:
  storageClassName: standard
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 20Gi
```

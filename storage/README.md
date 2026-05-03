# Storage — ShopOS

Persistent volume providers for stateful workloads. Cloud-managed disks (EBS, PD, Azure Disk)
are still preferred where available; this directory covers self-hosted and bare-metal options.

## Layout

| Subdir | Tool | Role |
|---|---|---|
| [longhorn/](longhorn/) | Longhorn 1.7 | Distributed RWO block storage from local node disks. Default for non-cloud clusters. Built-in snapshots + S3 backups. |
| [rook-ceph/](rook-ceph/) | Rook-Ceph v1.15 | Block + file (CephFS RWX) + S3-compatible object via single operator. For workloads needing RWX (image-processing-service, data-export-service). |

## Usage

```bash
# Default storage class (Longhorn)
kubectl get sc

# Provision RWX volume via CephFS
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: shared-images
  namespace: content
spec:
  accessModes: [ReadWriteMany]
  storageClassName: rook-cephfs
  resources: { requests: { storage: 100Gi } }
EOF
```

## Related

- Backups: [`kubernetes/velero/`](../kubernetes/velero/)
- MinIO (S3 object storage for assets / Velero / Loki / etc.): see [`databases/`](../databases/) for schema, deployed via Helm
- Cloud-native PV providers used in cloud clusters: AWS EBS CSI, GCP Persistent Disk CSI, Azure Disk CSI

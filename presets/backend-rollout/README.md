# KT Backend Rollout Preset（金丝雀发布标准）

> 适用：面向客户/核心链路的 KT 服务。内部小工具继续用普通 Deployment。

## 组成

| 文件 | 作用 |
|------|------|
| `rollout.yaml` | `kind: Rollout`，2 副本基线 + canary 阶梯（1 canary → 5% → 25% → 50% → 100%） |
| `analysis-template.yaml` | `AnalysisTemplate kt-canary-health-gate`：Job 探测新版本 `/healthz` + `/readyz`，全 200 自动放量，失败自动 abort 回滚 |

## 铁律（从 ktapi-file-istio 僵尸灰度 12 天的事故沉淀）

1. **任何 canary 必须有自动 promote 机制**（Analysis 或限时 pause），禁止纯人工 `pause: {}`
2. 新版本先起 **1 个 canary pod**，不许一上来双跑全量
3. 每档放量前过健康门；失败自动 abort，无需人工回滚
4. canary 期间的 PDB 以 stable 副本为准（canaryScale 另算）

## 接入步骤

1. 复制两个 yaml 到 `deploy/k8s-ack-prod/`，替换 `__SERVICE__ __NAMESPACE__ __IMAGE__ __PORT__`
2. 从 Deployment 迁移时：把原 Deployment 的 containers/env/volumes 搬进 rollout.yaml 的 `spec.template`，删除原 Deployment
3. Service 无需改动（selector 不变，stable+canary 共享）
4. kustomization 加入两个文件，推送 gitops 分支，Argo 同步
5. 验证：`kubectl -n argo-rollouts get rollout <svc>`

## 后续升级（指标门）

集群内落地独立 Prometheus 后，把 `analysis-template.yaml` 的 Job provider 换成：

```yaml
provider:
  prometheus:
    address: http://prometheus:9090
    query: |
      sum(rate(istio_requests_total{
        destination_workload="__SERVICE__",
        response_code=~"5.."
      }[2m]))
```

阈值：`successCondition: result[0] < 1`（每分钟 5xx < 1 次）。

## 回滚演练（每次迁移后必做）

```bash
# 1. 推一个带明显问题的 tag 到 gitops（如坏端口）
# 2. 观察 Rollout 到第一档健康门失败 → 自动 abort
kubectl -n <ns> get rollout <svc> -w
# 3. 确认 stable 全程 100% 接流量，公网探测不断
# 4. 恢复正确 tag，确认完整走到 100%
```

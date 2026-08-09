# 前端 preset（已迁移到 kt-frontend）

> **本目录不再放前端 app 骨架。** 前端 starter 的唯一家是 **`KTAIorg/kt-frontend`**。
> `kt-repo-template` 仅保留通用仓库治理（constitution/specs/ADR/CI）。

## 按产品类型选 starter（详见 `frontend-stack` skill v2）

| 产品类型 | 栈 | starter |
|---|---|---|
| 后台管理系统 | Refine + Ant Design | `kt-frontend/starters/admin-refine-antd` |
| 新产品 / 官网 | shadcn/ui + Tailwind | `kt-frontend/starters/web-shadcn`（规划中） |
| 移动 / TG WebApp | Vite + shadcn | `kt-frontend/starters/webapp-mobile`（规划中） |

判断口诀：**确定是后台管理系统 → Refine+antd（吃现成）；新产品/定制界面 → shadcn。**

## 认证
一律对接 kt-identity；面向客户优先 Telegram 一键登录。

## 决策
- `adr/0002-frontend-stack.md`（shadcn 线）
- `adr/0003-admin-stack-refine-antd.md`（后台 Refine+antd 线）

---
kind: plan
id: PLAN-2026-0001
status: review
owner: doudou
created: 2026-08-09
parent: SPEC-2026-0001
goal: null
services: [kt-agent-web, kt-specs]
contracts: []
---

# Plan: SpecDeck：kt-agent-web 需求文档浏览器

对应需求：`requirements/specdeck-kt-agent-web.md`（SPEC-2026-0001）。

## 影响范围

| 仓库 | 模块 | 变更 |
| --- | --- | --- |
| kt-specs | `.github/workflows/` | 新增索引生成 job：main 合并时产出 `spec-index.json` 并发布静态托管 |
| kt-agent-web | `src/components/specs/` | 新增 `SpecDeck.tsx`（+ 子组件），App 入口挂 Deck 开关 |
| kt-agent-web | `src/surfaceAdapter/` | 新增 `fetchSpecIndex()`（MVP 读静态文件，P1 换 core API 不改组件） |

## 不改范围

- GoalDeck/TodoDeck 现有逻辑不动；只抽取共用钻取原语（若抽取成本高于复制，MVP 允许复制，记入技术债）。
- kt CLI 与 frontmatter 契约不变。
- Backstage 不动（P1 才接 catalog）。

## 技术设计

### 分层模型（两轴一面）

```text
动机轴   goal    组织目标系统节点（GoalDeck 目标树，非本仓文档），spec 经 goal 字段挂入
         spec    kind: spec（顶层需求，goal 指向目标 id）
执行轴   plan    kind: plan（parent 指向 spec）
         tasks   kind: tasks（parent 指向 plan）
         验收    kind: checklist / evidence（parent 指向 tasks/plan）
视角轴   模块    按 frontmatter services 聚合（MVP）；P1 换 Backstage catalog
```

动机轴的边来自 frontmatter `goal` 字段（回指组织目标系统节点 id，事实源在
org 服务 `/org/goals`，SpecDeck 与 GoalDeck 共用同一棵目标树）；执行轴的边
来自 `parent`；`services` 只做视角聚合，不做层级父亲。阅读顺序即结晶的逆序：
先看目标（为什么做），再看 spec/plan/tasks/evidence（做了什么、做成了没）。
无 parent 的文档挂在所属 service 模块下；跨多 service 的文档在每个模块下
都出现（索引里记多份引用，正文只存一份）。未填 goal 的文档在目标层归入
「未挂目标」区，保证树中可达。

### spec-index.json schema（跨仓库合同）

生产方 kt-specs CI，消费方 kt-agent-web。变更须先改本节并同步两侧。

```jsonc
{
  "version": 1,
  "generated_at": "2026-08-09T14:00:00Z",   // 索引生成时间，UI 展示
  "source_commit": "abc1234",                 // kt-specs main SHA
  "modules": [
    {
      "id": "kt-agent-web",                   // service 名
      "title": "kt-agent-web",
      "docs": ["SPEC-2026-0001"]              // 挂在该模块下的顶层文档 id
    }
  ],
  "documents": [
    {
      "id": "SPEC-2026-0001",
      "kind": "spec",
      "status": "review",
      "title": "SpecDeck：kt-agent-web 需求文档浏览器",
      "owner": "doudou",
      "created": "2026-08-09",
      "parent": null,
      "goal": null,                            // 组织目标系统节点 id，未挂为 null
      "services": ["kt-agent-web", "kt-specs"],
      "path": "requirements/specdeck-kt-agent-web.md",  // 全文拉取用
      "excerpt": "KT Spec Platform 已落地……",            // 首段摘要 ≤200 字
      "children": ["PLAN-2026-0001"]          // 由 parent 反解
    }
  ]
}
```

规则：

- `children` 由 CI 从各文档 `parent` 反解生成，消费方不再自行组树。
- `goal` 原样透传 frontmatter，由消费方与组织目标树（`/org/goals`）联编；
  索引不内嵌目标数据，目标事实源在 org 服务。
- `path` 指向 raw 文档地址前缀 + 相对路径（托管 base URL 由消费方配置），
  焦点舱打开文档时按需拉全文，索引保持轻量。
- schema version 字段用于未来不兼容变更的协商。

### 索引生成（kt-specs CI）

在 Validate Specs workflow 之外新增 `publish-index` job：

- 触发：push 到 main（即文档 PR 合并后）。
- 步骤：跑校验器（复用 `.github/validator`）→ 生成索引 → 上传静态托管。
- 生成器与校验器同仓同模块（Go，`.github/spectools/`），解析 frontmatter
  逻辑共用 meta 包，避免第三份实现漂移。
- 托管决策（2026-08-09，**决策 B**）：新增 `KTAIorg/kt-spec-runtime`
  带鉴权数据面——runtime 经 GitHub tarball API 直接拉取 kt-specs 文档、
  进程内构建索引（TTL 刷新）；SpecDeck 客户端携带 KT Identity bearer
  读取 `/api/v1/specs/index|doc`。选择 B 而非 OSS/CDN 公开发布的原因：
  需求文档属内部信息，不应对公网可见，与组织全私有仓库的安全姿态一致。
- publish-index job 保留为 artifact 留痕（审计/调试），不再是前端数据源。

### SpecDeck 组件设计（kt-agent-web）

复用 GoalDeck 三栏骨架：

- 左栏：模块列表（L0）+ kind/status 过滤器。
- 中央：可缩放 outline 树。Enter 钻入（节点变根，面包屑压栈），
  Esc 回上层——与 GoalDeck 同款交互原语。
- 右坞焦点舱：文档元信息条（id/kind/status/owner/created/parent 链接）
  + `MarkdownBody` 渲染全文 + 子文档列表（继续钻取入口）。

状态色点映射（对齐 GoalDeck STATE_META 风格）：

| status | 视觉 |
| --- | --- |
| draft | 灰点 |
| review | 琥珀点 |
| approved | 青点 |
| implementing | 呼吸动画 |
| shipped | 绿点 |
| rejected / cancelled | 暗灰删除线 |

数据入口 `fetchSpecIndex()` 封装在 surfaceAdapter，MVP 实现为 fetch
静态 JSON；P1 切换 core API 时只改该函数签名内部。

## 数据模型

无新增数据库/存储。索引为无状态派生物，随时可由 main 重建。

## API / Contract

- `spec-index.json`（见上，唯一跨仓库合同）。
- P1 预留：core surface 增加 `GET /specs/tree`，返回同 schema，
  消费方零改动。

## 灰度和回滚

- SpecDeck 入口以 feature flag 控制（App 级开关），默认关闭，内部试用后放开。
- 回滚 = 关 flag；索引是静态派生物，删除无副作用；CI job 可独立 revert。

## 测试计划

| 类型 | 命令或方式 | 目标 |
| --- | --- | --- |
| 索引生成器单测 | `go test ./...`（kt-specs .github/indexer） | schema 合法性、parent 反解、多 service 挂载 |
| 契约一致性 | 用真实 main 文档生成索引，对照本 schema 逐字段核对 | 生产/消费两侧字段一致 |
| 前端组件测试 | vitest/单测（按 kt-agent-web 现行测试设施） | 钻取/回退/过滤/色点 |
| 手工冒烟 | 本地起 kt-agent-web 指向本地索引 | Enter/Esc/焦点舱渲染本 spec 文档 |
| dogfood | 合并后在 SpecDeck 中找到 SPEC-2026-0001 | 端到端闭环 |

## 迁移计划

无存量迁移。MVP 上线后，各仓 `.specs/` 的文档仍不进中央索引
（索引只覆盖 kt-specs 中央仓）；跨仓聚合属 P2，届时由 indexer
扩展数据源，schema 不变（documents 增加 repo 字段，version 升到 2）。

## 风险和取舍

- 摘要（excerpt）截取策略简单取首段，可能不够信息量——接受，MVP 不引入 NLP。
- 索引生成器与 validator 两个 Go 程序共仓，frontmatter 解析逻辑必须抽公共包，
  否则出现第三份实现（契约已明令禁止漂移）。
- 静态托管选型未定是本设计唯一悬而未决项；已给出 Releases 退化方案兜底。

# kt-specs — KT 跨仓库需求与设计文档中央仓

KT 规格驱动开发（spec-driven）的中央文档仓：**跨仓库的需求文档与设计文档
在这里统一管理**，仓库内部的需求/设计仍留在各业务仓的 `.specs/` 目录。

> 方法论事实源：[`KTAIorg/kt-spec-driven-playbook`](https://github.com/KTAIorg/kt-spec-driven-playbook)
> 字段契约：[`docs/methodology/spec-frontmatter-schema.md`](https://github.com/KTAIorg/kt-spec-driven-playbook/blob/main/docs/methodology/spec-frontmatter-schema.md)
> 平台 RFC：[`KTAIorg/kt-backstage` `.specs/kt-spec-platform/`](https://github.com/KTAIorg/kt-backstage/tree/main/.specs/kt-spec-platform)

## 目录结构

```text
requirements/   跨仓库需求文档（kind: spec），平铺存放
designs/        跨仓库设计文档（kind: plan），parent 指向需求 id
adr/            本仓自身的架构决策记录
```

## 使用方式（唯一程序化写入口：kt CLI）

```bash
# 新增需求（自动分配 id，生成带 frontmatter 的模板）
kt spec new spec --title "需求标题" --owner doudou --dir requirements --flat

# 新增设计文档（parent 指向需求 id）
kt spec new plan --title "方案标题" --owner doudou --parent SPEC-2026-0001 --dir designs --flat

# 校验 schema / id 唯一性 / 状态转移合法性
kt spec validate requirements designs

# 查询
kt spec list --status approved --json

# 状态流转（只能沿状态机前进）
kt spec status SPEC-2026-0001 review
```

## 状态机（PR 合并驱动）

```text
draft → review → approved → implementing → shipped
            ↘ rejected
任何非终态 → cancelled
```

- 评审 = PR review；**PR 合并即状态转移到 approved**（合并前把 status 改为目标态）。
- CI 会校验状态转移合法性：禁止回退与跨级跳转。
- 代码 PR 引用需求 id 时，该需求必须处于 approved 或更后状态。

## 门禁

- `Validate Specs` workflow：每个 PR 校验 frontmatter schema、id 唯一性、
  parent 引用与状态转移合法性。校验器是 `.github/validator/` 内的镜像实现，
  规则与 `kt-cli` 的 `kt spec validate` 一致（契约事实源在 playbook）；
  组织级跨仓库凭据就绪后切回直接调用 `kt spec validate`。
- main 为保护分支，只能经 PR 合入；`validate` 为 required check。

## 与仓库内 `.specs/` 的分工

| 场景 | 存放位置 |
| --- | --- |
| 跨多个仓库/服务的需求与设计 | 本仓 `requirements/`、`designs/` |
| 单仓库内部功能 | 该仓库 `.specs/<feature>/` |
| 跨仓库引用 | frontmatter `parent`/正文引用带仓库名前缀，如 `kt-specs:SPEC-2026-0001` |

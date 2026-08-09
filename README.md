<!--
═══════════════════════════════════════════════════════════════════
  用此模板创建新仓库后,请完成初始化,然后删除本注释块:
  1. 全局替换 <项目名称> 与 TODO 占位
  2. 填写 constitution.md(仓库定位/事实源/不可破坏边界)
  3. 填写 docs/(至少 domain-glossary + development-guide)
  4. 把 .specs/_example 与 contracts/_example.md 改成真实功能,或删除
  5. 编辑 .github/CODEOWNERS(指定审核团队)
  6. 按技术栈裁剪 .github/workflows/ci.yml 与 .gitignore
  7. 确认 .github/copilot-instructions.md 符合本仓库语言与审查规则
  8. 确认 LICENSE 与 SECURITY.md 联系方式
  9. 删除本注释块
═══════════════════════════════════════════════════════════════════
-->

# <项目名称>

> 一句话说明本仓库负责什么。TODO

## 简介

TODO: 背景、目标用户、核心能力、与其他仓库的关系。

## 快速开始

```bash
# TODO: 安装依赖、运行、测试
```

详见 [`docs/development-guide.md`](docs/development-guide.md)。

## 文档索引

本仓库遵循 **KT 规格驱动开发(spec-driven)** 方法论 —— AI Agent 与开发者在动代码前必须先读文档。

| 文档 | 作用 |
| --- | --- |
| [`AGENTS.md`](AGENTS.md) | AI Agent 工作规则(必读) |
| [`constitution.md`](constitution.md) | 仓库定位、事实源、不可破坏边界 |
| [`docs/domain-glossary.md`](docs/domain-glossary.md) | 领域术语与统一命名 |
| [`docs/cross-repo-boundaries.md`](docs/cross-repo-boundaries.md) | 跨仓库职责边界 |
| [`docs/development-guide.md`](docs/development-guide.md) | 本地开发、目录、命令 |
| [`docs/testing-strategy.md`](docs/testing-strategy.md) | 测试层级与验收 |
| [`docs/security-privacy.md`](docs/security-privacy.md) | 安全、脱敏、权限 |
| [`docs/observability.md`](docs/observability.md) | 日志、trace、metrics |
| [`docs/release-runbook.md`](docs/release-runbook.md) | 发布、灰度、回滚 |
| [`docs/debugging-runbook.md`](docs/debugging-runbook.md) | 排查与诊断 |
| [`contracts/`](contracts/) | API/事件/schema 合同 |
| [`adr/`](adr/) | 架构决策记录 |
| [`.specs/`](.specs/) | 功能级 spec/plan/tasks/checklist/evidence |

## 功能开发流程

每个重要功能在 `.specs/<feature>/` 下建立:

```
spec.md      做什么、不做什么、验收是什么
  ↓
plan.md      怎么做、影响哪些模块、如何回滚
  ↓
tasks.md     拆成可执行任务
  ↓
checklist.md 开发前/提交前/发布前检查
  ↓
evidence.md  验证命令、结果、残留风险
```

参考 [`.specs/_example/`](.specs/_example/)。

## 贡献

见 [`CONTRIBUTING.md`](CONTRIBUTING.md)。代码审查由 Copilot 自动进行,输出简体中文。

## 语言与 AI 协作约定

- 文档、PR 描述、提交信息、变更说明、发布说明默认使用简体中文。
- 代码标识符、API 路径、配置 key、环境变量、命令和第三方产品名称保留英文。
- 仓库级 Copilot 指令位于 [`.github/copilot-instructions.md`](.github/copilot-instructions.md)，新仓库从模板创建后应保留并按领域补充。

## 方法论来源

完整规范: [`KTAIorg/kt-spec-driven-playbook`](https://github.com/KTAIorg/kt-spec-driven-playbook)

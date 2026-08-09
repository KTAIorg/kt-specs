# AGENTS.md

本文定义 AI Agent 在本仓库工作的规则。

## 开始前必读

AI Agent 接到任务后必须先阅读：

1. `README.md`
2. `constitution.md`
3. `docs/domain-glossary.md`
4. `docs/cross-repo-boundaries.md`
5. 相关 `.specs/<feature>/spec.md`
6. 相关 `.specs/<feature>/plan.md`
7. 相关 `contracts/*.md`
8. 相关 `adr/*.md`
9. 最后才阅读代码。

## 工作规则

- 文档、PR 描述、提交信息、变更说明默认使用简体中文；代码标识符、API 路径、配置 key、环境变量和命令保留英文。
- 不清楚边界时，先补文档或提出问题。
- 涉及跨仓库字段时，先更新 contract。
- 涉及架构取舍时，新增或更新 ADR。
- 涉及高风险操作时，更新 checklist 和 evidence。
- 不要绕过本仓库的事实源。
- 不要修改与任务无关的文件。

## 允许改动范围

TODO

## 禁止改动范围

TODO

## 常用命令

```bash
TODO
```

## 提交前检查

- [ ] 相关 spec 已更新。
- [ ] 相关 contract 已更新。
- [ ] 相关 ADR 已更新或确认不需要。
- [ ] 测试或校验已执行。
- [ ] evidence 已补充。

## 高风险规则

TODO

## 方法论来源

本仓库遵循 KT 规格驱动开发(spec-driven)方法论,完整规范见 [`KTAIorg/kt-spec-driven-playbook`](https://github.com/KTAIorg/kt-spec-driven-playbook):

- 文档体系: `docs/methodology/ai-agent-required-documents.md`
- 方法论总览: `docs/methodology/kt-spec-driven-development.md`

代码审查由 Copilot 自动进行并使用简体中文,规则见 `.github/copilot-instructions.md`。

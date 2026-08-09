# 贡献指南

## 开始之前

动代码前请先阅读 [`AGENTS.md`](AGENTS.md) 与 [`constitution.md`](constitution.md),并确认相关 `.specs/<feature>/` 文档已就绪。文档缺失时先补文档草案,再进入实现。

## 分支规范

| 前缀 | 用途 |
| --- | --- |
| `feature/*` | 新功能 |
| `fix/*` | Bug 修复 |
| `chore/*` | 构建、依赖、杂项维护 |

主干: `main`(生产) ← `develop`(集成,如使用)。预发使用 `pre` 分支(如使用)。请勿在历史归档分支(如 `legacy-main`)上继续开发。

## 提交信息

- 默认使用简体中文，用清晰的祈使句描述改动(会显示在 GitHub Actions 运行名)。
- 推荐 Conventional Commits: `feat:` `fix:` `chore:` `docs:` `refactor:` `test:`。

## 语言约定

- 文档、PR 描述、提交信息、变更说明、发布说明默认使用简体中文。
- 代码标识符、API 路径、配置 key、环境变量、命令和第三方产品名称保留英文。
- 如果上游协议或开源社区要求英文，可在相关文件中保留英文，但面向 KT 内部协作的说明仍优先中文。

## Pull Request 检查清单

- [ ] 关联了 issue 或 `.specs/<feature>`。
- [ ] 通过 CI(lint + test + 安全扫描)。
- [ ] 增补或更新了测试。
- [ ] 更新了相关文档(spec / plan / contract / ADR)。
- [ ] 高风险改动更新了 `evidence.md`。
- [ ] PR 描述说明了影响范围与回滚方式。

## 代码风格

- 遵循 [`.editorconfig`](.editorconfig)。
- 运行本仓库的 `make fmt` / `make lint`(如有)。
- 代码审查由 Copilot 自动进行,输出简体中文,规则见 [`.github/copilot-instructions.md`](.github/copilot-instructions.md)。

## 安全

发现漏洞请勿公开提 issue,流程见 [`SECURITY.md`](SECURITY.md)。

# 安全政策

## 报告漏洞

**请勿在公开 issue 中报告安全漏洞。**

请通过邮件联系: `security@TODO`(TODO: 替换为实际安全联系邮箱),并提供:

- 漏洞描述与影响范围
- 复现步骤
- 受影响的版本 / 分支
- 可能的修复建议

我们会在收到后尽快响应并跟进修复。

## 支持的版本

| 版本 / 分支 | 是否支持 |
| --- | --- |
| `main` | ✅ |
| 历史归档分支 | ❌ |

## 安全实践

- 使用 Infisical / GitHub Secrets 管理敏感凭证,严禁硬编码密钥与令牌。
- 实际 `.env` 不入版本控制,仅提交 `.env.example`。
- 依赖更新由 Dependabot 自动跟进,见 [`.github/dependabot.yml`](.github/dependabot.yml)。
- 代码安全扫描由 CodeQL 执行,见 [`.github/workflows/codeql.yml`](.github/workflows/codeql.yml)。
- 数据分类与处理要求见 [`docs/security-privacy.md`](docs/security-privacy.md)。

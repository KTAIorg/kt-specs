# Constitution

## 仓库定位

本仓库负责：

- 跨仓库需求文档（requirements/）与设计文档（designs/）的唯一存放地。
- 规格文档状态机的执行载体：PR 合并驱动状态转移。
- 为 kt CLI `spec` 子命令与 CI 门禁提供被校验的文档事实。

本仓库不负责：

- 单仓库内部功能的规格文档（留在各业务仓 `.specs/`）。
- frontmatter 字段契约的定义（事实源在 kt-spec-driven-playbook）。
- 方法论本身的演进（改 playbook，不改这里）。

## 事实源

本仓库的事实源包括：

- `requirements/*.md`、`designs/*.md` 的 frontmatter 与正文。
- 本仓 PR 历史（状态转移与评审记录）。

本仓库依赖的外部事实源包括：

- `kt-spec-driven-playbook` `docs/methodology/spec-frontmatter-schema.md`（字段契约）。
- `kt-cli` `internal/spec`（契约的校验实现）。

## 不可破坏边界

- 不得绕过 `kt spec` CLI 直接改写 frontmatter 的 status 字段。
- 不得引入与 playbook 契约不一致的字段或状态。
- 不得把仓库内局部需求搬进本仓制造双轨。
- main 分支只接受经 PR 与 CI 校验的变更。

## 文档优先级

AI 和开发者在改文档前必须先阅读：

1. `README.md`
2. `constitution.md`
3. playbook 的 frontmatter 契约
4. 相关 `adr/*`

## 安全和审计

- 需求文档不得包含密钥、token、客户个人数据原文；涉及敏感内容只写引用。
- 全部变更经 Git 历史可审计；状态转移必须可追溯到 PR。

## 验收原则

- 新文档必须通过 `kt spec validate`。
- 需求必须可回答：背景、目标、非目标、验收标准。
- 设计文档必须通过 `parent` 关联到需求 id。

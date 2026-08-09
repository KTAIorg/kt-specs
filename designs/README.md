# designs/

跨仓库设计文档（kind: plan），平铺存放；frontmatter `parent` 必须指向
已存在的需求 id（如 `SPEC-2026-0001`）。

新增：

```bash
kt spec new plan --title "方案标题" --owner <你> --parent SPEC-2026-0001 --dir designs --flat
```

状态流转与 requirements/ 一致（draft → review → approved → implementing → shipped）。

# requirements/

跨仓库需求文档（kind: spec），平铺存放，文件名 = 标题 slug。

新增：

```bash
kt spec new spec --title "需求标题" --owner <你> --dir requirements --flat
```

评审流程：补全 TODO 字段 → `kt spec status <ID> review` → 开 PR →
评审通过后将 status 改为 `approved` 再合并（合并即状态转移生效）。

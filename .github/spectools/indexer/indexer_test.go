package indexer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const specA = `---
kind: spec
id: SPEC-2026-0001
status: review
owner: doudou
created: 2026-08-09
parent: null
services: [svc-x, svc-y]
contracts: []
---

# Spec: 需求A

这是需求A的首段摘要，描述背景与动机。

## 背景

后续段落不应进入摘要。
`

const planB = `---
kind: plan
id: PLAN-2026-0001
status: draft
owner: doudou
created: 2026-08-09
parent: SPEC-2026-0001
services: [svc-x]
contracts: []
---

# Plan: 设计B

设计摘要。
`

const specC = `---
kind: spec
id: SPEC-2026-0002
status: draft
owner: doudou
created: 2026-08-09
parent: null
services: []
contracts: []
---

# Spec: 无服务归属
`

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func fixtureDir(t *testing.T) string {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "requirements", "a.md"), specA)
	writeFile(t, filepath.Join(dir, "designs", "b.md"), planB)
	writeFile(t, filepath.Join(dir, "requirements", "c.md"), specC)
	writeFile(t, filepath.Join(dir, "requirements", "nofront.md"), "# 无 frontmatter\n")
	return dir
}

func TestBuild(t *testing.T) {
	idx, err := Build([]string{fixtureDir(t)}, "abc1234")
	if err != nil {
		t.Fatal(err)
	}
	if idx.Version != 1 || idx.SourceCommit != "abc1234" {
		t.Fatalf("顶层字段异常: %+v", idx)
	}

	// 模块聚合：多 service 顶层文档在每个模块都出现；子文档不在模块层
	moduleDocs := map[string][]string{}
	for _, m := range idx.Modules {
		moduleDocs[m.ID] = m.Docs
	}
	if got := moduleDocs["svc-x"]; len(got) != 1 || got[0] != "SPEC-2026-0001" {
		t.Fatalf("svc-x 应只含 SPEC-2026-0001（PLAN 有 parent 不列模块层）: %v", got)
	}
	if got := moduleDocs["svc-y"]; len(got) != 1 || got[0] != "SPEC-2026-0001" {
		t.Fatalf("svc-y 应含 SPEC-2026-0001: %v", got)
	}
	if got := moduleDocs[unclassifiedID]; len(got) != 1 || got[0] != "SPEC-2026-0002" {
		t.Fatalf("无 services 文档应进 unclassified: %v", got)
	}

	byID := map[string]Document{}
	for _, d := range idx.Documents {
		byID[d.ID] = d
	}
	if len(byID) != 3 {
		t.Fatalf("应索引 3 篇文档（无 frontmatter 的跳过）: %d", len(byID))
	}

	a := byID["SPEC-2026-0001"]
	if a.Title != "Spec: 需求A" {
		t.Fatalf("标题提取异常: %q", a.Title)
	}
	if !strings.Contains(a.Excerpt, "首段摘要") || strings.Contains(a.Excerpt, "后续段落") {
		t.Fatalf("摘要应取首段: %q", a.Excerpt)
	}
	if a.Parent != nil {
		t.Fatal("顶层文档 parent 应为 null")
	}
	if len(a.Children) != 1 || a.Children[0] != "PLAN-2026-0001" {
		t.Fatalf("children 反解异常: %v", a.Children)
	}

	b := byID["PLAN-2026-0001"]
	if b.Parent == nil || *b.Parent != "SPEC-2026-0001" {
		t.Fatalf("parent 解析异常: %+v", b.Parent)
	}
}

func TestBuildRejectsDuplicateID(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "requirements", "a.md"), specA)
	writeFile(t, filepath.Join(dir, "requirements", "dup.md"),
		strings.Replace(specA, "# Spec: 需求A", "# Spec: 重复", 1))
	if _, err := Build([]string{dir}, ""); err == nil {
		t.Fatal("重复 id 应报错")
	}
}

func TestExtractExcerptSkipsFenceAndTruncates(t *testing.T) {
	body := "# 标题\n\n```\n代码块不应进入摘要\n```\n\n" + strings.Repeat("长", 300) + "\n"
	got := ExtractExcerpt(body)
	if strings.Contains(got, "代码块") {
		t.Fatalf("不应包含代码围栏内容: %q", got)
	}
	if len([]rune(got)) > excerptRunes {
		t.Fatalf("摘要应截断到 %d runes: %d", excerptRunes, len([]rune(got)))
	}
}

func TestExtractTitleFallback(t *testing.T) {
	if got := ExtractTitle("没有标题的正文", "SPEC-2026-0099"); got != "SPEC-2026-0099" {
		t.Fatalf("无标题应回退 id: %q", got)
	}
}

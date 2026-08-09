package meta

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validDoc = `---
kind: spec
id: SPEC-2026-0001
status: draft
owner: doudou
created: 2026-08-09
parent: null
services: [kt-fans-api]
contracts: []
---

# Spec: 示例

正文段落。
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

func mustParse(t *testing.T, content string) Doc {
	t.Helper()
	path := filepath.Join(t.TempDir(), "doc.md")
	writeFile(t, path, content)
	doc, err := ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func TestExtractFrontmatter(t *testing.T) {
	yamlText, ok := ExtractFrontmatter(validDoc)
	if !ok || !strings.Contains(yamlText, "kind: spec") {
		t.Fatalf("frontmatter 提取失败: ok=%v", ok)
	}
	if _, ok := ExtractFrontmatter("# 无 frontmatter\n"); ok {
		t.Fatal("无 frontmatter 文档不应提取成功")
	}
}

func TestParseFileBody(t *testing.T) {
	doc := mustParse(t, validDoc)
	if doc.Meta == nil || doc.Meta.ID != "SPEC-2026-0001" {
		t.Fatalf("meta 解析异常: %+v", doc.Meta)
	}
	if !strings.Contains(doc.Body, "# Spec: 示例") {
		t.Fatalf("body 应保留正文: %q", doc.Body)
	}
	if strings.Contains(doc.Body, "kind: spec") {
		t.Fatal("body 不应包含 frontmatter")
	}
}

func TestMetaValidateValid(t *testing.T) {
	doc := mustParse(t, validDoc)
	if issues := doc.Meta.Validate(); len(issues) != 0 {
		t.Fatalf("合法元数据不应有问题: %v", issues)
	}
}

func TestMetaValidateInvalid(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*Meta)
		wantSub string
	}{
		{"kind 非法", func(m *Meta) { m.Kind = "wiki" }, "kind"},
		{"id 缺失", func(m *Meta) { m.ID = "" }, "id 缺失"},
		{"id 格式非法", func(m *Meta) { m.ID = "spec-1" }, "id 格式"},
		{"id 前缀不匹配", func(m *Meta) { m.ID = "PLAN-2026-0001" }, "前缀"},
		{"status 非法", func(m *Meta) { m.Status = "done" }, "status"},
		{"owner 缺失", func(m *Meta) { m.Owner = "" }, "owner"},
		{"created 非法", func(m *Meta) { m.Created = "2026/08/09" }, "created"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			doc := mustParse(t, validDoc)
			tc.mutate(doc.Meta)
			issues := doc.Meta.Validate()
			found := false
			for _, msg := range issues {
				if strings.Contains(msg, tc.wantSub) {
					found = true
				}
			}
			if !found {
				t.Fatalf("期望包含 %q 的问题，实际: %v", tc.wantSub, issues)
			}
		})
	}
}

func TestValidTransition(t *testing.T) {
	legal := [][2]string{
		{"draft", "review"}, {"draft", "cancelled"},
		{"review", "approved"}, {"review", "rejected"}, {"review", "cancelled"},
		{"approved", "implementing"}, {"approved", "cancelled"},
		{"implementing", "shipped"},
	}
	for _, pair := range legal {
		if !ValidTransition(pair[0], pair[1]) {
			t.Errorf("应允许 %s -> %s", pair[0], pair[1])
		}
	}
	illegal := [][2]string{
		{"draft", "approved"},
		{"review", "draft"},
		{"shipped", "implementing"},
		{"rejected", "review"},
		{"cancelled", "draft"},
		{"implementing", "approved"},
	}
	for _, pair := range illegal {
		if ValidTransition(pair[0], pair[1]) {
			t.Errorf("应拒绝 %s -> %s", pair[0], pair[1])
		}
	}
}

func TestMarkdownFilesSkipsTemplates(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "requirements", "a.md"), validDoc)
	writeFile(t, filepath.Join(dir, "templates", "spec.md"), validDoc)
	writeFile(t, filepath.Join(dir, "_example", "spec.md"), validDoc)
	files, err := MarkdownFiles([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || !strings.Contains(files[0], "requirements") {
		t.Fatalf("应只保留 requirements/a.md: %v", files)
	}
}

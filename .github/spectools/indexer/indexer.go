// Package indexer 生成 spec-index.json：kt-agent-web SpecDeck 的数据源。
//
// schema 合同定义在 kt-specs designs/specdeck-kt-agent-web.md（PLAN-2026-0001）
// 「spec-index.json schema」一节；合同变更必须先改 PLAN 并同步消费方。
package indexer

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/KTAIorg/kt-specs/spectools/meta"
)

const (
	indexVersion   = 1
	excerptRunes   = 200
	unclassifiedID = "unclassified"
)

// Module 是 L0 模块层节点（MVP 按 frontmatter services 聚合）。
type Module struct {
	ID    string   `json:"id"`
	Title string   `json:"title"`
	Docs  []string `json:"docs"` // 挂在该模块下的顶层文档 id（parent 为空）
}

// Document 是一篇文档的索引条目。
type Document struct {
	ID       string   `json:"id"`
	Kind     string   `json:"kind"`
	Status   string   `json:"status"`
	Title    string   `json:"title"`
	Owner    string   `json:"owner"`
	Created  string   `json:"created"`
	Parent   *string  `json:"parent"` // 无 parent 时为 null
	Services []string `json:"services"`
	Path     string   `json:"path"`     // 相对仓库根的路径，全文拉取用
	Excerpt  string   `json:"excerpt"`  // 首段摘要 ≤200 runes
	Children []string `json:"children"` // 由 parent 反解
}

// Index 是 spec-index.json 顶层结构。
type Index struct {
	Version      int        `json:"version"`
	GeneratedAt  string     `json:"generated_at"`
	SourceCommit string     `json:"source_commit"`
	Modules      []Module   `json:"modules"`
	Documents    []Document `json:"documents"`
}

// Build 扫描给定目录生成索引，路径规范化为相对扫描根的路径。
func Build(paths []string, sourceCommit string) (*Index, error) {
	files := map[string]string{}
	for _, root := range paths {
		list, err := meta.MarkdownFiles([]string{root})
		if err != nil {
			return nil, err
		}
		info, err := os.Stat(root)
		if err != nil {
			return nil, err
		}
		// 相对路径保留扫描根目录名（requirements/x.md 与 designs/x.md 不冲突）
		base := root
		if info.IsDir() {
			if d := filepath.Dir(strings.TrimRight(root, "/")); d != "" {
				base = d
			}
		}
		for _, f := range list {
			rel := filepath.Base(f)
			if r, err := filepath.Rel(base, f); err == nil && r != "" && !strings.HasPrefix(r, "..") {
				rel = r
			}
			data, err := os.ReadFile(f)
			if err != nil {
				return nil, err
			}
			files[filepath.ToSlash(rel)] = string(data)
		}
	}
	return BuildFromMap(files, sourceCommit)
}

// BuildFromMap 从「仓库相对路径 -> 内容」集合构建索引；内存数据源
// （如 spec-runtime 的 GitHub 拉取结果）直接使用，不落盘。
func BuildFromMap(files map[string]string, sourceCommit string) (*Index, error) {
	keys := make([]string, 0, len(files))
	for k := range files {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	byID := map[string]*Document{}
	var docs []*Document
	for _, path := range keys {
		doc, err := meta.ParseContent(path, files[path])
		if err != nil || doc.Meta == nil || doc.Meta.ID == "" {
			continue
		}
		if _, dup := byID[doc.Meta.ID]; dup {
			return nil, fmt.Errorf("id %q 重复，索引无法生成（先通过 validate 门禁）", doc.Meta.ID)
		}
		parent := normalizeParent(doc.Meta.Parent)
		d := &Document{
			ID:       doc.Meta.ID,
			Kind:     doc.Meta.Kind,
			Status:   doc.Meta.Status,
			Title:    ExtractTitle(doc.Body, doc.Meta.ID),
			Owner:    doc.Meta.Owner,
			Created:  doc.Meta.Created,
			Parent:   parent,
			Services: append([]string{}, doc.Meta.Services...),
			Path:     path,
			Excerpt:  ExtractExcerpt(doc.Body),
			Children: []string{},
		}
		byID[d.ID] = d
		docs = append(docs, d)
	}

	// children 由 parent 反解；parent 不在索引集合内（跨仓库引用）则忽略
	for _, d := range docs {
		if d.Parent != nil {
			if parent, ok := byID[*d.Parent]; ok {
				parent.Children = append(parent.Children, d.ID)
			}
		}
	}

	moduleDocs := map[string][]string{}
	for _, d := range docs {
		if d.Parent != nil {
			continue // 子文档经 children 下钻，不在模块层重复列出
		}
		services := d.Services
		if len(services) == 0 {
			services = []string{unclassifiedID}
		}
		for _, svc := range services {
			moduleDocs[svc] = append(moduleDocs[svc], d.ID)
		}
	}
	var modules []Module
	for id, docIDs := range moduleDocs {
		sort.Strings(docIDs)
		modules = append(modules, Module{ID: id, Title: id, Docs: docIDs})
	}
	sort.Slice(modules, func(i, j int) bool { return modules[i].ID < modules[j].ID })

	outDocs := make([]Document, 0, len(docs))
	for _, d := range docs {
		sort.Strings(d.Children)
		outDocs = append(outDocs, *d)
	}
	sort.Slice(outDocs, func(i, j int) bool { return outDocs[i].ID < outDocs[j].ID })

	return &Index{
		Version:      indexVersion,
		GeneratedAt:  time.Now().UTC().Format(time.RFC3339),
		SourceCommit: sourceCommit,
		Modules:      modules,
		Documents:    outDocs,
	}, nil
}

func normalizeParent(parent string) *string {
	if parent == "" || parent == "null" {
		return nil
	}
	return &parent
}

// ExtractTitle 取正文第一个一级标题；没有则回退为文档 id。
func ExtractTitle(body, fallback string) string {
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return fallback
}

// ExtractExcerpt 取标题之后的第一个正文段落（跳过标题/表格/代码围栏/引用），
// 截断到 excerptRunes 个字符。
func ExtractExcerpt(body string) string {
	inFence := false
	seenTitle := false
	var paragraph []string
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		if strings.HasPrefix(trimmed, "# ") {
			seenTitle = true
			continue
		}
		if trimmed == "" || strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "|") || strings.HasPrefix(trimmed, ">") ||
			strings.HasPrefix(trimmed, "-") {
			if len(paragraph) > 0 {
				break
			}
			continue
		}
		if !seenTitle {
			continue
		}
		paragraph = append(paragraph, trimmed)
		// 只取第一个段落
		break
	}
	text := strings.Join(paragraph, " ")
	runes := []rune(text)
	if len(runes) > excerptRunes {
		return string(runes[:excerptRunes])
	}
	return text
}

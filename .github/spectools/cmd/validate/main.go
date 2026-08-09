// validate 是 kt-specs 仓 CI 门禁：校验 frontmatter schema、id 唯一性、
// parent 引用与状态转移合法性。规则实现见 meta 包（与 kt-cli 镜像）。
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/KTAIorg/kt-specs/spectools/meta"
)

type issue struct {
	path string
	msg  string
}

func main() {
	paths := os.Args[1:]
	if len(paths) == 0 {
		paths = []string{"."}
	}
	files, err := meta.MarkdownFiles(paths)
	if err != nil {
		fmt.Fprintln(os.Stderr, "错误:", err)
		os.Exit(2)
	}

	type docRef struct {
		path string
		m    *meta.Meta
	}
	var docs []docRef
	var errors, warnings []issue
	validated := 0

	for _, f := range files {
		doc, err := meta.ParseFile(f)
		if err != nil {
			errors = append(errors, issue{f, err.Error()})
			continue
		}
		if doc.Meta == nil {
			if meta.IsSpecLocation(f) && !strings.EqualFold(fileName(f), "README.md") {
				warnings = append(warnings, issue{f, "缺少 frontmatter（迁移窗口内仅告警）"})
			}
			continue
		}
		docs = append(docs, docRef{f, doc.Meta})
		validated++
		for _, msg := range doc.Meta.Validate() {
			errors = append(errors, issue{f, msg})
		}
	}

	seen := map[string]string{}
	for _, d := range docs {
		if d.m.ID == "" {
			continue
		}
		if prev, dup := seen[d.m.ID]; dup {
			errors = append(errors, issue{d.path, fmt.Sprintf("id %q 与 %s 重复", d.m.ID, prev)})
		} else {
			seen[d.m.ID] = d.path
		}
	}
	for _, d := range docs {
		p := d.m.Parent
		if p == "" || p == "null" || strings.Contains(p, ":") {
			continue
		}
		if _, ok := seen[p]; !ok {
			warnings = append(warnings, issue{d.path, fmt.Sprintf("parent %q 在本次扫描范围内不存在", p)})
		}
	}
	for _, d := range docs {
		old, ok := meta.HeadStatus(d.path)
		if !ok || old == "" || old == d.m.Status {
			continue
		}
		if !meta.ValidTransition(old, d.m.Status) {
			errors = append(errors, issue{d.path, fmt.Sprintf("状态转移非法: %s -> %s", old, d.m.Status)})
		}
	}

	for _, w := range warnings {
		fmt.Printf("[warning] %s: %s\n", w.path, w.msg)
	}
	for _, e := range errors {
		fmt.Printf("[error] %s: %s\n", e.path, e.msg)
	}
	fmt.Printf("扫描 %d 个文件，校验 %d 篇文档，错误 %d，告警 %d\n", len(files), validated, len(errors), len(warnings))
	if len(errors) > 0 {
		os.Exit(22)
	}
}

func fileName(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}

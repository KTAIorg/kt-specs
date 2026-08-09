// validator 是 kt-specs 仓 CI 用的 frontmatter 校验器。
//
// 它是 KTAIorg/kt-cli internal/spec 校验规则的**镜像实现**：字段事实源在
// kt-spec-driven-playbook docs/methodology/spec-frontmatter-schema.md。
// 字段或状态机变更时必须同步两处。长期方向：组织级跨仓库凭据就绪后，
// 改回直接调用 `kt spec validate`（见 kt-backstage SPEC-2026-0001 plan）。
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	validKinds    = []string{"spec", "plan", "tasks", "checklist", "evidence", "adr", "contract"}
	validStatuses = []string{"draft", "review", "approved", "implementing", "shipped", "rejected", "cancelled"}
	kindPrefix    = map[string]string{
		"spec": "SPEC", "plan": "PLAN", "tasks": "TASKS", "checklist": "CHECK",
		"evidence": "EVID", "adr": "ADR", "contract": "CONTRACT",
	}
	transitions = map[string][]string{
		"draft":        {"review", "cancelled"},
		"review":       {"approved", "rejected", "cancelled"},
		"approved":     {"implementing", "cancelled"},
		"implementing": {"shipped"},
	}
	idRegexp   = regexp.MustCompile(`^([A-Z]+)-(\d{4})-(\d{4})$`)
	dateLayout = "2006-01-02"
	skipDirs   = map[string]bool{
		".git": true, "node_modules": true, ".yarn": true, "vendor": true,
		"templates": true, "_example": true,
	}
)

type meta struct {
	Kind      string   `yaml:"kind"`
	ID        string   `yaml:"id"`
	Status    string   `yaml:"status"`
	Owner     string   `yaml:"owner"`
	Created   string   `yaml:"created"`
	Parent    string   `yaml:"parent"`
	Services  []string `yaml:"services"`
	Contracts []string `yaml:"contracts"`
}

type issue struct {
	path string
	msg  string
}

func main() {
	paths := os.Args[1:]
	if len(paths) == 0 {
		paths = []string{"."}
	}
	files, err := collectMarkdown(paths)
	if err != nil {
		fmt.Fprintln(os.Stderr, "错误:", err)
		os.Exit(2)
	}

	type doc struct {
		path string
		meta *meta
	}
	var docs []doc
	var errors, warnings []issue
	validated := 0

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			errors = append(errors, issue{f, err.Error()})
			continue
		}
		yamlText, ok := extractFrontmatter(string(data))
		if !ok {
			if isSpecLocation(f) && !strings.EqualFold(filepath.Base(f), "README.md") {
				warnings = append(warnings, issue{f, "缺少 frontmatter（迁移窗口内仅告警）"})
			}
			continue
		}
		var m meta
		if err := yaml.Unmarshal([]byte(yamlText), &m); err != nil {
			errors = append(errors, issue{f, fmt.Sprintf("frontmatter YAML 解析失败: %v", err)})
			continue
		}
		docs = append(docs, doc{f, &m})
		validated++
		for _, msg := range validateMeta(&m) {
			errors = append(errors, issue{f, msg})
		}
	}

	seen := map[string]string{}
	for _, d := range docs {
		if d.meta.ID == "" {
			continue
		}
		if prev, dup := seen[d.meta.ID]; dup {
			errors = append(errors, issue{d.path, fmt.Sprintf("id %q 与 %s 重复", d.meta.ID, prev)})
		} else {
			seen[d.meta.ID] = d.path
		}
	}
	for _, d := range docs {
		p := d.meta.Parent
		if p == "" || p == "null" || strings.Contains(p, ":") {
			continue
		}
		if _, ok := seen[p]; !ok {
			warnings = append(warnings, issue{d.path, fmt.Sprintf("parent %q 在本次扫描范围内不存在", p)})
		}
	}
	for _, d := range docs {
		old, ok := headStatus(d.path)
		if !ok || old == "" || old == d.meta.Status {
			continue
		}
		if !validTransition(old, d.meta.Status) {
			errors = append(errors, issue{d.path, fmt.Sprintf("状态转移非法: %s -> %s", old, d.meta.Status)})
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

func extractFrontmatter(content string) (string, bool) {
	if !strings.HasPrefix(content, "---\n") {
		return "", false
	}
	rest := content[4:]
	idx := strings.Index(rest, "\n---\n")
	if idx < 0 {
		return "", false
	}
	return rest[:idx], true
}

func validateMeta(m *meta) []string {
	var issues []string
	if !contains(validKinds, m.Kind) {
		issues = append(issues, fmt.Sprintf("kind 非法或缺失: %q（合法值: %s）", m.Kind, strings.Join(validKinds, "/")))
	}
	if m.ID == "" {
		issues = append(issues, "id 缺失")
	} else if !idRegexp.MatchString(m.ID) {
		issues = append(issues, fmt.Sprintf("id 格式非法: %q（应为 PREFIX-YYYY-NNNN）", m.ID))
	} else if prefix, ok := kindPrefix[m.Kind]; ok {
		if got := idRegexp.FindStringSubmatch(m.ID)[1]; got != prefix {
			issues = append(issues, fmt.Sprintf("id 前缀 %q 与 kind %q 不匹配（应为 %s）", got, m.Kind, prefix))
		}
	}
	if !contains(validStatuses, m.Status) {
		issues = append(issues, fmt.Sprintf("status 非法或缺失: %q（合法值: %s）", m.Status, strings.Join(validStatuses, "/")))
	}
	if strings.TrimSpace(m.Owner) == "" {
		issues = append(issues, "owner 缺失")
	}
	if m.Created == "" {
		issues = append(issues, "created 缺失")
	} else if _, err := time.Parse(dateLayout, m.Created); err != nil {
		issues = append(issues, fmt.Sprintf("created 日期格式非法: %q（应为 YYYY-MM-DD）", m.Created))
	}
	return issues
}

func validTransition(from, to string) bool {
	for _, next := range transitions[from] {
		if next == to {
			return true
		}
	}
	return false
}

func isSpecLocation(path string) bool {
	for _, seg := range strings.Split(filepath.ToSlash(path), "/") {
		switch seg {
		case ".specs", "requirements", "designs", "contracts", "adr":
			return true
		}
	}
	switch filepath.Base(path) {
	case "spec.md", "plan.md", "tasks.md", "checklist.md", "evidence.md":
		return true
	}
	return false
}

// headStatus 读取 git HEAD 版本的 status；不可用时跳过（与 kt CLI 行为一致）。
func headStatus(path string) (string, bool) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", false
	}
	dir := filepath.Dir(abs)
	out, err := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", false
	}
	root := strings.TrimSpace(string(out))
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return "", false
	}
	data, err := exec.Command("git", "-C", root, "show", "HEAD:"+filepath.ToSlash(rel)).Output()
	if err != nil {
		return "", false
	}
	yamlText, ok := extractFrontmatter(string(data))
	if !ok {
		return "", false
	}
	var m meta
	if err := yaml.Unmarshal([]byte(yamlText), &m); err != nil {
		return "", false
	}
	return m.Status, true
}

func collectMarkdown(paths []string) ([]string, error) {
	var files []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, fmt.Errorf("无法访问 %s: %w", p, err)
		}
		if !info.IsDir() {
			if strings.HasSuffix(strings.ToLower(p), ".md") {
				files = append(files, p)
			}
			continue
		}
		err = filepath.WalkDir(p, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if skipDirs[d.Name()] {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Strings(files)
	return files, nil
}

func contains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

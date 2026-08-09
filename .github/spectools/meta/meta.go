// Package meta 是 kt-specs 仓内 frontmatter 解析与校验规则的共享实现，
// 供 cmd/validate（CI 门禁）与 indexer（索引生成）复用。
//
// 字段事实源在 KTAIorg/kt-spec-driven-playbook
// docs/methodology/spec-frontmatter-schema.md；规则与 KTAIorg/kt-cli
// internal/spec 保持一致（镜像关系，契约变更须同步）。
package meta

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

// 合法枚举值，与契约文档一致。
var (
	ValidKinds    = []string{"spec", "plan", "tasks", "checklist", "evidence", "adr", "contract"}
	ValidStatuses = []string{"draft", "review", "approved", "implementing", "shipped", "rejected", "cancelled"}
)

// KindPrefix 是 id 前缀与 kind 的映射。
var KindPrefix = map[string]string{
	"spec":      "SPEC",
	"plan":      "PLAN",
	"tasks":     "TASKS",
	"checklist": "CHECK",
	"evidence":  "EVID",
	"adr":       "ADR",
	"contract":  "CONTRACT",
}

// transitions 是状态机的全部合法转移。终态（shipped/rejected/cancelled）无出边。
var transitions = map[string][]string{
	"draft":        {"review", "cancelled"},
	"review":       {"approved", "rejected", "cancelled"},
	"approved":     {"implementing", "cancelled"},
	"implementing": {"shipped"},
}

var (
	idRegexp   = regexp.MustCompile(`^([A-Z]+)-(\d{4})-(\d{4})$`)
	dateLayout = "2006-01-02"
)

// skipDirNames 在遍历时整体跳过：依赖目录、产物目录，
// 以及刻意保留占位符的模板与示例目录。
var skipDirNames = map[string]bool{
	".git":         true,
	"node_modules": true,
	".yarn":        true,
	"vendor":       true,
	"templates":    true,
	"_example":     true,
}

// Meta 是 frontmatter 机器可读元数据。
type Meta struct {
	Kind      string   `yaml:"kind"`
	ID        string   `yaml:"id"`
	Status    string   `yaml:"status"`
	Owner     string   `yaml:"owner"`
	Created   string   `yaml:"created"`
	Parent    string   `yaml:"parent"`
	Goal      *int     `yaml:"goal"`
	Services  []string `yaml:"services"`
	Contracts []string `yaml:"contracts"`
}

// Doc 是一篇带（或不带）frontmatter 的 Markdown 文档。
type Doc struct {
	Path string
	Meta *Meta  // nil 表示没有 frontmatter
	Body string // frontmatter 之后的正文
}

// ExtractFrontmatter 从 Markdown 内容中提取 YAML frontmatter 文本。
func ExtractFrontmatter(content string) (string, bool) {
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

// ParseFile 读取并解析一篇文档。
func ParseFile(path string) (Doc, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Doc{}, err
	}
	return ParseContent(path, string(data))
}

// ParseContent 从内容字符串解析一篇文档（内存数据源用）。
func ParseContent(path, content string) (Doc, error) {
	doc := Doc{Path: path}
	yamlText, ok := ExtractFrontmatter(content)
	if !ok {
		doc.Body = content
		return doc, nil
	}
	var m Meta
	if err := yaml.Unmarshal([]byte(yamlText), &m); err != nil {
		return doc, fmt.Errorf("frontmatter YAML 解析失败: %w", err)
	}
	doc.Meta = &m
	doc.Body = content[len("---\n")+len(yamlText)+len("\n---\n"):]
	return doc, nil
}

// Validate 检查元数据字段合法性，返回全部问题描述。
func (m *Meta) Validate() []string {
	var issues []string
	if !Contains(ValidKinds, m.Kind) {
		issues = append(issues, fmt.Sprintf("kind 非法或缺失: %q（合法值: %s）", m.Kind, strings.Join(ValidKinds, "/")))
	}
	if m.ID == "" {
		issues = append(issues, "id 缺失")
	} else if !idRegexp.MatchString(m.ID) {
		issues = append(issues, fmt.Sprintf("id 格式非法: %q（应为 PREFIX-YYYY-NNNN）", m.ID))
	} else if prefix, ok := KindPrefix[m.Kind]; ok {
		if got := idRegexp.FindStringSubmatch(m.ID)[1]; got != prefix {
			issues = append(issues, fmt.Sprintf("id 前缀 %q 与 kind %q 不匹配（应为 %s）", got, m.Kind, prefix))
		}
	}
	if !Contains(ValidStatuses, m.Status) {
		issues = append(issues, fmt.Sprintf("status 非法或缺失: %q（合法值: %s）", m.Status, strings.Join(ValidStatuses, "/")))
	}
	if strings.TrimSpace(m.Owner) == "" {
		issues = append(issues, "owner 缺失")
	}
	if m.Created == "" {
		issues = append(issues, "created 缺失")
	} else if _, err := time.Parse(dateLayout, m.Created); err != nil {
		issues = append(issues, fmt.Sprintf("created 日期格式非法: %q（应为 YYYY-MM-DD）", m.Created))
	}
	if m.Goal != nil && *m.Goal <= 0 {
		issues = append(issues, fmt.Sprintf("goal 格式非法: %d（应为正整数，指向组织目标系统节点 id）", *m.Goal))
	}
	for _, s := range m.Services {
		if strings.TrimSpace(s) == "" {
			issues = append(issues, "services 含空值")
			break
		}
	}
	for _, c := range m.Contracts {
		if strings.TrimSpace(c) == "" {
			issues = append(issues, "contracts 含空值")
			break
		}
	}
	return issues
}

// ValidTransition 判断状态转移是否合法。
func ValidTransition(from, to string) bool {
	for _, next := range transitions[from] {
		if next == to {
			return true
		}
	}
	return false
}

// IsSpecLocation 判断路径是否属于规格文档约定位置。
func IsSpecLocation(path string) bool {
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

// HeadStatus 读取 git 基线版本中该文档的 status 字段；不可用时返回 ok=false。
//
// 基线选择：工作区内容与 HEAD 不一致时以 HEAD 为基线（本地未提交场景）；
// 两者一致时（CI 检出后工作区恒等于 HEAD，检查会恒为空操作）回退 HEAD^1——
// pull_request 的合并提交上即 base 侧，保证状态转移门禁在 CI 中真实生效。
func HeadStatus(path string) (string, bool) {
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
	relSlash := filepath.ToSlash(rel)
	working, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	ref := "HEAD"
	headBlob, ok := gitShow(root, relSlash, "HEAD")
	if !ok {
		return "", false // 新增文档，无既往状态可比对
	}
	if headBlob == string(working) {
		if _, ok2 := gitShow(root, relSlash, "HEAD^1"); !ok2 {
			return "", false
		}
		ref = "HEAD^1"
	}
	blob, ok := gitShow(root, relSlash, ref)
	if !ok {
		return "", false
	}
	yamlText, ok := ExtractFrontmatter(blob)
	if !ok {
		return "", false
	}
	var m Meta
	if err := yaml.Unmarshal([]byte(yamlText), &m); err != nil {
		return "", false
	}
	return m.Status, true
}

func gitShow(root, rel, ref string) (string, bool) {
	data, err := exec.Command("git", "-C", root, "show", ref+":"+rel).Output()
	if err != nil {
		return "", false
	}
	return string(data), true
}

// MarkdownFiles 递归收集路径下的 Markdown 文件，跳过依赖/模板/示例目录。
func MarkdownFiles(paths []string) ([]string, error) {
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
				if skipDirNames[d.Name()] {
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

// Contains 判断字符串切片是否包含值。
func Contains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

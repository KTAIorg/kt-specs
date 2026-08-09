// indexer 生成 spec-index.json（SpecDeck 数据源）。
//
// 用法：
//
//	go run ./cmd/indexer --out spec-index.json requirements designs
//
// SOURCE_COMMIT 环境变量或 --commit 记录生成时的 main SHA。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/KTAIorg/kt-specs/spectools/indexer"
)

func main() {
	out := flag.String("out", "spec-index.json", "输出文件路径")
	commit := flag.String("commit", os.Getenv("SOURCE_COMMIT"), "源 commit SHA")
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "用法: indexer [--out FILE] [--commit SHA] DIR...")
		os.Exit(2)
	}
	idx, err := indexer.Build(flag.Args(), *commit)
	if err != nil {
		fmt.Fprintln(os.Stderr, "索引生成失败:", err)
		os.Exit(1)
	}
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "序列化失败:", err)
		os.Exit(1)
	}
	data = append(data, '\n')
	if err := os.WriteFile(*out, data, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "写入失败:", err)
		os.Exit(1)
	}
	fmt.Printf("已生成 %s：%d 个模块，%d 篇文档（commit=%s）\n", *out, len(idx.Modules), len(idx.Documents), idx.SourceCommit)
}

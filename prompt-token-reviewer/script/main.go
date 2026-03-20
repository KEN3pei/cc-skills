package main

import (
	"fmt"
	"os"

	"prompt-reviewer/internal/history"
)

func main() {
	prompts, err := history.ReadRandomPrompts(10)
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: 会話履歴の読み込みに失敗しました: %v\n", err)
		os.Exit(1)
	}

	for i, p := range prompts {
		fmt.Printf("%d. %s\n", i+1, p.FullText)
	}
}

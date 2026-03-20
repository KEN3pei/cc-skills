package history

import (
	"bufio"
	"encoding/json"
	"errors"
	"math/rand"
	"os"
)

// minDisplayLength はフィルタ条件：display のルーン数がこれ以上なら有効とみなす。
// これによって/exitのようなコマンド系や簡単すぎるプロンプトはレビュー対象から外している。
const minDisplayLength = 15

// Prompt は会話履歴から抽出した1件のユーザープロンプトを表す。
type Prompt struct {
	Display  string
	FullText string
}

// historyEntry は history.jsonl の1行を表す。
type historyEntry struct {
	Display        string                   `json:"display"`
	PastedContents map[string]pastedContent `json:"pastedContents"`
}

type pastedContent struct {
	ID      int    `json:"id"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

// ReadRandomPrompts は ~/.claude/history.jsonl からランダムに最大 limit 件のプロンプトを返す。
func ReadRandomPrompts(limit int) ([]Prompt, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	path := home + "/.claude/history.jsonl"
	return ReadPromptsFromFile(path, limit)
}

// ReadPromptsFromFile は指定パスの JSONL ファイルからランダムに最大 limit 件のプロンプトを返す。
// テスト可能にするため path を引数に取る。
func ReadPromptsFromFile(path string, limit int) ([]Prompt, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var filtered []Prompt
	scanner := bufio.NewScanner(f)
	// 長い行にも対応できるよう 1MB バッファを設定する
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		var entry historyEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			// 不正な JSON 行はスキップする
			continue
		}

		pastedText := collectPastedText(entry.PastedContents)
		hasMeaningfulPasted := pastedText != ""
		hasLongDisplay := len([]rune(entry.Display)) >= minDisplayLength

		if !hasLongDisplay && !hasMeaningfulPasted {
			continue
		}

		fullText := entry.Display
		if pastedText != "" {
			fullText += "\n" + pastedText
		}

		filtered = append(filtered, Prompt{
			Display:  entry.Display,
			FullText: fullText,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(filtered) == 0 {
		return nil, errors.New("有効なプロンプトが見つかりませんでした（全エントリが短すぎるか空です）")
	}

	rand.Shuffle(len(filtered), func(i, j int) {
		filtered[i], filtered[j] = filtered[j], filtered[i]
	})

	n := limit
	if len(filtered) < n {
		n = len(filtered)
	}
	return filtered[:n], nil
}

// collectPastedText は pastedContents から type=="text" のコンテンツを結合して返す。
// history.jsonlの履歴からtype=imageのものはレビューできないのでここで対象から外している。
func collectPastedText(contents map[string]pastedContent) string {
	var result string
	for _, c := range contents {
		if c.Type == "text" && c.Content != "" {
			if result != "" {
				result += "\n"
			}
			result += c.Content
		}
	}
	return result
}

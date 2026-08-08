---
name: android-tabs
description: AndroidのChromeで開いているタブを取得し、技術系のものだけAIが要約してMarkdownファイルに保存する。確認後に対象タブを閉じる。
---

# android-tabs スキル

AndroidのChromeタブを取得し、技術系のものだけ要約してMarkdownファイルに保存する。確認後に対象タブを閉じる。

## 前提条件

`android-tabs` CLIツールがインストールされていること。

- リポジトリ: `~/Apps/android-tabs/`
- セットアップ: `~/Apps/android-tabs/README.md` を参照
- Chromeアプリをフォアグラウンドで起動しておくこと（バックグラウンドだとDevTools Protocolに接続できない）

## 使い方

```
/android-tabs [保存先パス]
```

例:
```
/android-tabs ~/Desktop/tabs-summary.md
/android-tabs ~/Documents/tabs/2026-08-06.md
```

保存先パスを省略した場合は、`~/.claude/skills/android-tabs/config` の `OUTPUT_DIR` を参照する。そちらも未設定の場合は `~/chrome-tabs-{YYYY-MM-DD}.md` に保存する。

## 実行フロー

以下の手順で実行すること。各ステップで失敗した場合はエラー内容をユーザーに伝えて停止する。

### ステップ0: 保存先パスの決定

以下の優先順位で保存先パスを決定する:

1. `$ARGUMENTS` が指定されていればそれを使う
2. `~/.claude/skills/android-tabs/config` が存在すれば `OUTPUT_DIR` を読み取り、`{OUTPUT_DIR}/tabs-{YYYY-MM-DD}.md` を使う
3. いずれもなければ `~/chrome-tabs-{YYYY-MM-DD}.md` を使う

configファイルの読み取りは以下のBashコマンドで行う:
```bash
grep '^OUTPUT_DIR=' ~/.claude/skills/android-tabs/config | cut -d= -f2
```

### ステップ1: タブ一覧取得

```bash
android-tabs list
```

- 失敗した場合（exit code != 0）: stderrの内容をユーザーに伝えて停止
- 結果はJSON配列: `[{"id":"...","title":"...","url":"..."},...]`

### ステップ2: 技術系タブのフィルタリング

取得したタブのタイトルとURLを見て、以下の基準で技術系か判断する:

**技術系と判断するもの**
- プログラミング・開発・インフラ・クラウド・AI・セキュリティ・データベースに関連するURL/タイトル
- GitHub, Zenn, Qiita, Stack Overflow, 公式ドキュメント, 技術ブログ など
- 論文・技術仕様書・RFC など

**技術系でないと判断するもの**
- ニュース・SNS・ショッピング・動画・ゲーム・グルメ など

### ステップ3: 各タブの要約

技術系と判断したタブについて、以下を実行する:

1. WebFetchでURLの本文を取得する
2. 取得できた場合 → 本文を基に要約を生成
3. 取得できなかった場合（ログイン必須・動的サイトなど）→ タイトルとURLのみで要約を生成（その旨を記載）

要約フォーマット:
```
サマリー1行（何について書かれているかを端的に）
- 箇条書きポイント1
- 箇条書きポイント2
- 箇条書きポイント3
```

### ステップ4: Markdownファイルへの書き出し

以下のフォーマットでファイルを書き出す:

```markdown
# Android Chrome タブ要約 YYYY-MM-DD

取得タブ数: XX件 / 技術系: XX件

---

## {タイトル}

URL: {URL}
取得日時: YYYY-MM-DD HH:MM

{サマリー1行}

- {ポイント1}
- {ポイント2}
- {ポイント3}

---

## {タイトル}

...
```

ファイルを書き出したら、保存先パスをユーザーに伝える。

### ステップ5: タブを閉じる確認

ユーザーに以下を伝える:
- 技術系タブの一覧（番号・タイトル・URL）
- 「全タブを閉じます。除外したいタブのIDをカンマ区切りで入力してください（なければそのままEnter）」

AskUserQuestionで確認を取る。

入力されたIDを除外した残りのIDで以下を実行:

```bash
android-tabs close --ids {id1},{id2},...
```

### ステップ6: 完了報告

- 閉じたタブ数
- 保存したファイルパス
- （あれば）除外したタブ一覧

を報告して終了。

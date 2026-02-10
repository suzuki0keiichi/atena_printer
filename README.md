# atena_printer - 年賀状宛名印刷ツール

Google Spreadsheet から住所録を読み込み、はがき宛名面の PDF を生成する Go 製 CLI ツール。

## セットアップ

### 1. Google Cloud の準備

1. [Google Cloud Console](https://console.cloud.google.com/) でプロジェクトを作成
2. Google Sheets API を有効化
3. サービスアカウントを作成し、JSON 鍵ファイルをダウンロード
4. スプレッドシートをサービスアカウントのメールアドレスに共有（編集権限）

### 2. スプレッドシートの準備

シート名「住所録」（config で変更可能）に以下のヘッダ行を作成:

| 姓 | 名 | 連名 | 敬称 | 郵便番号 | 住所1 | 住所2 | 2026送 | 2026受 | 2026喪中 |
|----|-----|------|------|---------|-------|-------|--------|--------|----------|

- **姓** / **名**: 宛先の姓名
- **連名**: 連名がある場合（カンマ or 読点区切りで複数可。例: `花子、一郎`）
- **敬称**: 空欄なら「様」が自動適用
- **郵便番号**: ハイフン有無どちらでも可（例: `100-0001` / `1000001`）
- **住所1**: 都道府県から番地まで
- **住所2**: 建物名・部屋番号など（任意）
- **YYYY送 / YYYY受 / YYYY喪中**: 年ごとのステータス列。何か入力すれば有効と判定（推奨: ○）

年が変わったら `2027送`, `2027受`, `2027喪中` のように列を追加していく。

### 3. フォントの準備

日本語 TrueType フォント（.ttf）が必要。以下は無料の例:

- **IPAex明朝** (ipaexm.ttf): https://moji.or.jp/ipafont/
- **Noto Serif JP**: Google Fonts からダウンロード

### 4. 設定ファイル

`config.example.json` をコピーして `config.json` を作成:

```bash
cp config.example.json config.json
```

各項目を自分の環境に合わせて編集:

```json
{
  "spreadsheet_id": "スプレッドシートのURL中のID",
  "sheet_name": "住所録",
  "credentials_file": "credentials.json",
  "font_file": "/path/to/ipaexm.ttf",
  "output_file": "nenga.pdf",
  "year": 2026,
  "sender": {
    "family_name": "山田",
    "given_name": "太郎",
    "postal_code": "1000001",
    "address1": "東京都千代田区千代田一丁目一番",
    "address2": ""
  }
}
```

### 5. ビルド

```bash
go build -o atena_printer .
```

## 使い方

### 宛名 PDF を生成

```bash
# 今年の未送付・非喪中の宛先だけ出力
./atena_printer generate

# 全件出力
./atena_printer generate -all

# 出力先を指定
./atena_printer generate -output nenga_2026.pdf

# 設定ファイルを指定
./atena_printer generate -config /path/to/config.json
```

生成された PDF をプリンタで印刷（はがきサイズ・等倍・フチなし推奨）。

### 住所一覧を確認

```bash
./atena_printer list
```

### 印刷済みを記録

```bash
# まず対象を確認（dry-run）
./atena_printer mark-sent -dry-run

# 実際にスプレッドシートに書き込み
./atena_printer mark-sent
```

対象の宛先の「YYYY送」列に ○ が記録される。

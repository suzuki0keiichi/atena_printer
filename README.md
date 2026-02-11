# atena_printer - 年賀状宛名印刷ツール

Google Spreadsheet から住所録を読み込み、はがき宛名面の PDF を生成する Go 製 CLI ツール。

## セットアップ

### 1. スプレッドシートの準備

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

#### モードA: 公開シート読み取り（Google Cloud不要）

Google Cloud 設定をせずに使う場合は、シートを「リンクを知っている全員が閲覧可」にする。
この場合、`generate` / `list` が利用できる。

#### モードB: 手動TSV読み取り（Google Cloud不要・非公開シート向け）

1. Google Spreadsheet で対象シートを開く
2. `ファイル` → `ダウンロード` → `タブ区切り値 (.tsv、現在のシート)`
3. 生成された `.tsv` をローカルに保存
4. `config.json` の `tsv_file` にそのパスを設定

この場合も `generate` / `list` が利用できる。

#### 書き込みモード（上級）

`mark-sent` でシート更新まで行う場合のみ、Google Cloud でサービスアカウントを用意する。

1. [Google Cloud Console](https://console.cloud.google.com/) でプロジェクトを作成
2. Google Sheets API を有効化
3. サービスアカウントを作成し、JSON 鍵ファイルをダウンロード
4. スプレッドシートをサービスアカウントのメールアドレスに共有（編集権限）

### 2. フォントの準備

日本語 TrueType フォント（.ttf）が必要。以下は無料の例:

- **IPAex明朝** (ipaexm.ttf): https://moji.or.jp/ipafont/
- **Noto Serif JP**: Google Fonts からダウンロード

#### 推奨プリセット

- 住所・氏名: `YujiSyuku-Regular.ttf`（毛筆感はあるが主張が強すぎない）
- 郵便番号: `Arial Unicode.ttf` などの機械系フォント

### 3. 設定ファイル

`config.example.json` をコピーして `config.json` を作成:

```bash
cp config.example.json config.json
```

各項目を自分の環境に合わせて編集:

```json
{
  "spreadsheet_id": "スプレッドシートのURL中のID",
  "sheet_name": "住所録",
  "credentials_file": "",
  "tsv_file": "",
  "font_file": "/path/to/YujiSyuku-Regular.ttf",
  "postal_font_file": "/path/to/Arial-Unicode.ttf",
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

- `tsv_file` が空でない場合: `tsv_file` を優先してローカルTSVから読み込む（`generate` / `list`）。
- `tsv_file` が空かつ `credentials_file` が空の場合: 公開シートから読み込む（`generate` / `list`）。
- `credentials_file` に JSON 鍵ファイルを指定した場合: 読み書き可能モード（`mark-sent`）が利用可能。
- `postal_font_file` は任意。設定すると郵便番号だけ別フォントにできる（未設定時は `font_file` を使用）。

### 4. ビルド

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
`mark-sent` は `tsv_file` 未使用かつ `credentials_file` 設定時のみ利用可能。

## 免責

- 本ツールの利用に伴う住所録データの取得・管理・保管・共有設定・運用は、利用者自身の責任で行ってください。
- 個人情報保護、法令順守、誤送付防止、データバックアップ等の実務上の責任は利用者にあります。
- 本ツールの利用または利用不能により生じた損害（データ消失、誤印刷、誤送付、逸失利益等）について、作者は責任を負いません。

## ライセンス

- ソフトウェア本体: BSD 3-Clause License（`LICENSE`）
- 同梱フォントなど第三者ライセンス: `THIRD_PARTY_LICENSES.md`

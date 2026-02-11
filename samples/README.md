# サンプルデータ

## 含まれるファイル

- `addresses.sample.tsv`: 住所録のサンプルTSV
- `config.sample.tsv.json`: TSV読み取りモード用の設定例（YujiSyuku推奨プリセット）
- `sample_output.pdf`: サンプルTSVから生成した出力例
- `fonts/*.ttf`: 比較用の毛筆フォント
- `config.font.*.json`: フォント比較用の設定ファイル
- `renders/*.pdf`: 各フォントで生成した比較結果

## 確認コマンド

```bash
go run . list -config samples/config.sample.tsv.json
go run . generate -config samples/config.sample.tsv.json -output samples/sample_output.pdf
```

## 毛筆フォント比較の生成

```bash
mkdir -p samples/renders
go run . generate -config samples/config.font.yujiboku.json
go run . generate -config samples/config.font.yujimai.json
go run . generate -config samples/config.font.yujisyuku.json
go run . generate -config samples/config.font.zenkurenaido.json
```

- 郵便番号は `postal_font_file` で機械系フォントを指定。
- 住所・氏名は `font_file` で毛筆フォントを指定。

## 備考

- `sample_output.pdf` は `2026送` が空の行のみ出力されるため、2件分のページが生成される。
- 全件を出す場合は `-all` を付ける。

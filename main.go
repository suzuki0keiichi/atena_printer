package main

import (
	"flag"
	"fmt"
	"os"

	"atena_printer/internal/config"
	"atena_printer/internal/model"
	"atena_printer/internal/pdf"
	"atena_printer/internal/sheets"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "generate":
		cmdGenerate(args)
	case "mark-sent":
		cmdMarkSent(args)
	case "list":
		cmdList(args)
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "不明なコマンド: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `年賀状宛名印刷ツール

使い方:
  atena_printer <command> [options]

コマンド:
  generate     宛名PDFを生成する
  mark-sent    印刷済みの宛先をスプレッドシートに記録する
  list         住所一覧とステータスを表示する
  help         この使い方を表示する

共通オプション:
  -config string  設定ファイルのパス (default: config.json)

generate オプション:
  -all           喪中・送付済みを含めて全件出力する
  -output string 出力ファイルパス (設定ファイルの値を上書き)

mark-sent オプション:
  -dry-run       実際には書き込まず対象を表示する
`)
}

func cmdGenerate(args []string) {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	configPath := fs.String("config", "config.json", "設定ファイルのパス")
	all := fs.Bool("all", false, "全件出力する")
	output := fs.String("output", "", "出力ファイルパス")
	fs.Parse(args)

	cfg, err := config.Load(*configPath)
	if err != nil {
		exitError(err)
	}
	if *output != "" {
		cfg.OutputFile = *output
	}

	client, err := sheets.New(cfg.CredentialsFile, cfg.SpreadsheetID, cfg.SheetName)
	if err != nil {
		exitError(err)
	}

	addresses, statuses, err := client.ReadAddresses(cfg.Year)
	if err != nil {
		exitError(err)
	}

	// フィルタリング
	var targets []model.Address
	for _, addr := range addresses {
		st := statuses[addr.Row]
		if !*all {
			if st.Sent || st.Mourning {
				continue
			}
		}
		targets = append(targets, addr)
	}

	if len(targets) == 0 {
		fmt.Println("出力対象の宛先がありません。")
		return
	}

	gen, err := pdf.NewGenerator(cfg.FontFile, cfg.Sender)
	if err != nil {
		exitError(err)
	}

	for _, addr := range targets {
		if err := gen.AddPage(addr); err != nil {
			exitError(fmt.Errorf("%s%s の処理中にエラー: %w", addr.FamilyName, addr.GivenName, err))
		}
	}

	if err := gen.Save(cfg.OutputFile); err != nil {
		exitError(fmt.Errorf("PDF の保存に失敗: %w", err))
	}

	fmt.Printf("PDF を生成しました: %s (%d件)\n", cfg.OutputFile, len(targets))
}

func cmdMarkSent(args []string) {
	fs := flag.NewFlagSet("mark-sent", flag.ExitOnError)
	configPath := fs.String("config", "config.json", "設定ファイルのパス")
	dryRun := fs.Bool("dry-run", false, "実際には書き込まず対象を表示する")
	fs.Parse(args)

	cfg, err := config.Load(*configPath)
	if err != nil {
		exitError(err)
	}

	client, err := sheets.New(cfg.CredentialsFile, cfg.SpreadsheetID, cfg.SheetName)
	if err != nil {
		exitError(err)
	}

	addresses, statuses, err := client.ReadAddresses(cfg.Year)
	if err != nil {
		exitError(err)
	}

	var rows []int
	for _, addr := range addresses {
		st := statuses[addr.Row]
		if !st.Sent && !st.Mourning {
			rows = append(rows, addr.Row)
			fmt.Printf("  %s %s (%s)\n", addr.FamilyName, addr.GivenName, addr.Address1)
		}
	}

	if len(rows) == 0 {
		fmt.Println("更新対象がありません。")
		return
	}

	if *dryRun {
		fmt.Printf("\n%d件が対象です (dry-run: 書き込みはしません)\n", len(rows))
		return
	}

	if err := client.MarkSent(cfg.Year, rows); err != nil {
		exitError(err)
	}

	fmt.Printf("\n%d件を %d送 = ○ に更新しました。\n", len(rows), cfg.Year)
}

func cmdList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	configPath := fs.String("config", "config.json", "設定ファイルのパス")
	fs.Parse(args)

	cfg, err := config.Load(*configPath)
	if err != nil {
		exitError(err)
	}

	client, err := sheets.New(cfg.CredentialsFile, cfg.SpreadsheetID, cfg.SheetName)
	if err != nil {
		exitError(err)
	}

	addresses, statuses, err := client.ReadAddresses(cfg.Year)
	if err != nil {
		exitError(err)
	}

	fmt.Printf("--- %d年 住所一覧 (%d件) ---\n", cfg.Year, len(addresses))
	for _, addr := range addresses {
		st := statuses[addr.Row]
		sentMark := " "
		recvMark := " "
		mournMark := " "
		if st.Sent {
			sentMark = "○"
		}
		if st.Received {
			recvMark = "○"
		}
		if st.Mourning {
			mournMark = "喪"
		}

		joint := ""
		if len(addr.JointNames) > 0 {
			joint = " ほか"
		}

		fmt.Printf("  [送:%s 受:%s %s] %s %s%s%s  〒%s %s%s\n",
			sentMark, recvMark, mournMark,
			addr.FamilyName, addr.GivenName, joint, addr.Honorific,
			formatPostalCode(addr.PostalCode),
			addr.Address1, addr.Address2)
	}
}

func formatPostalCode(code string) string {
	if len(code) == 7 {
		return code[:3] + "-" + code[3:]
	}
	return code
}

func exitError(err error) {
	fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
	os.Exit(1)
}

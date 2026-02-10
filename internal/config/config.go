package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Sender struct {
	FamilyName string `json:"family_name"`
	GivenName  string `json:"given_name"`
	PostalCode string `json:"postal_code"`
	Address1   string `json:"address1"`
	Address2   string `json:"address2"`
}

type Config struct {
	SpreadsheetID   string `json:"spreadsheet_id"`
	SheetName       string `json:"sheet_name"`
	CredentialsFile string `json:"credentials_file"`
	FontFile        string `json:"font_file"`
	OutputFile      string `json:"output_file"`
	Year            int    `json:"year"`
	Sender          Sender `json:"sender"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("設定ファイルを読み込めません: %w", err)
	}

	cfg := &Config{
		SheetName:  "住所録",
		OutputFile: "nenga.pdf",
		Year:       time.Now().Year(),
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("設定ファイルの形式が不正です: %w", err)
	}

	if cfg.SpreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheet_id が設定されていません")
	}
	if cfg.CredentialsFile == "" {
		return nil, fmt.Errorf("credentials_file が設定されていません")
	}
	if cfg.FontFile == "" {
		return nil, fmt.Errorf("font_file が設定されていません")
	}
	if cfg.Sender.FamilyName == "" {
		return nil, fmt.Errorf("sender.family_name が設定されていません")
	}

	return cfg, nil
}

package sheets

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"atena_printer/internal/model"
)

const sheetsAPIBase = "https://sheets.googleapis.com/v4/spreadsheets"

type Client struct {
	ts            *tokenSource
	httpClient    *http.Client
	spreadsheetID string
	sheetName     string
}

func New(credentialsFile, spreadsheetID, sheetName string) (*Client, error) {
	ts, err := newTokenSource(credentialsFile)
	if err != nil {
		return nil, err
	}
	return &Client{
		ts:            ts,
		httpClient:    &http.Client{},
		spreadsheetID: spreadsheetID,
		sheetName:     sheetName,
	}, nil
}

// ReadAddresses はスプレッドシートから住所一覧を読み込む。
func (c *Client) ReadAddresses(year int) ([]model.Address, map[int]model.YearStatus, error) {
	readRange := c.sheetName + "!A1:ZZ"
	values, err := c.getValues(readRange)
	if err != nil {
		return nil, nil, fmt.Errorf("スプレッドシートの読み込みに失敗: %w", err)
	}

	if len(values) < 2 {
		return nil, nil, fmt.Errorf("スプレッドシートにデータがありません")
	}

	header := values[0]
	colIdx := buildColumnIndex(header)
	yearCols := findYearColumns(header, year)

	var addresses []model.Address
	statuses := make(map[int]model.YearStatus)

	for i, row := range values[1:] {
		rowNum := i + 2 // 1-indexed, skip header

		familyName := getCell(row, colIdx["姓"])
		if familyName == "" {
			continue
		}

		postalCode := normalizePostalCode(getCell(row, colIdx["郵便番号"]))

		addr := model.Address{
			FamilyName: familyName,
			GivenName:  getCell(row, colIdx["名"]),
			JointNames: parseJointNames(getCell(row, colIdx["連名"])),
			Honorific:  getCell(row, colIdx["敬称"]),
			PostalCode: postalCode,
			Address1:   getCell(row, colIdx["住所1"]),
			Address2:   getCell(row, colIdx["住所2"]),
			Row:        rowNum,
		}
		if addr.Honorific == "" {
			addr.Honorific = "様"
		}

		addresses = append(addresses, addr)

		status := model.YearStatus{
			Sent:     isChecked(getCell(row, yearCols.sent)),
			Received: isChecked(getCell(row, yearCols.received)),
			Mourning: isChecked(getCell(row, yearCols.mourning)),
		}
		statuses[rowNum] = status
	}

	return addresses, statuses, nil
}

// MarkSent はスプレッドシートの指定行の「YYYY送」列に ○ を書き込む。
func (c *Client) MarkSent(year int, rows []int) error {
	headerValues, err := c.getValues(c.sheetName + "!1:1")
	if err != nil {
		return fmt.Errorf("ヘッダ行の読み込みに失敗: %w", err)
	}
	if len(headerValues) == 0 {
		return fmt.Errorf("ヘッダ行が空です")
	}

	header := headerValues[0]
	sentColName := fmt.Sprintf("%d送", year)
	sentColIdx := -1
	for i, h := range header {
		if h == sentColName {
			sentColIdx = i
			break
		}
	}
	if sentColIdx < 0 {
		return fmt.Errorf("列 '%s' が見つかりません。スプレッドシートに列を追加してください", sentColName)
	}

	colLetter := columnLetter(sentColIdx)

	var data []batchData
	for _, row := range rows {
		cellRange := fmt.Sprintf("%s!%s%d", c.sheetName, colLetter, row)
		data = append(data, batchData{
			Range:  cellRange,
			Values: [][]string{{"○"}},
		})
	}

	return c.batchUpdate(data)
}

// --- Sheets API raw HTTP ---

type valuesResponse struct {
	Values [][]string `json:"values"`
}

func (c *Client) getValues(readRange string) ([][]string, error) {
	token, err := c.ts.getToken()
	if err != nil {
		return nil, err
	}

	u := fmt.Sprintf("%s/%s/values/%s",
		sheetsAPIBase,
		c.spreadsheetID,
		url.PathEscape(readRange))

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API エラー (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var result valuesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("API 応答の解析に失敗: %w", err)
	}

	return result.Values, nil
}

type batchData struct {
	Range  string     `json:"range"`
	Values [][]string `json:"values"`
}

type batchUpdateRequest struct {
	ValueInputOption string      `json:"valueInputOption"`
	Data             []batchData `json:"data"`
}

func (c *Client) batchUpdate(data []batchData) error {
	token, err := c.ts.getToken()
	if err != nil {
		return err
	}

	u := fmt.Sprintf("%s/%s/values:batchUpdate", sheetsAPIBase, c.spreadsheetID)

	reqBody := batchUpdateRequest{
		ValueInputOption: "USER_ENTERED",
		Data:             data,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", u, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API エラー (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// --- helpers ---

type yearColumns struct {
	sent     int
	received int
	mourning int
}

func buildColumnIndex(header []string) map[string]int {
	idx := make(map[string]int)
	for i, h := range header {
		idx[h] = i
	}
	return idx
}

func findYearColumns(header []string, year int) yearColumns {
	yc := yearColumns{sent: -1, received: -1, mourning: -1}
	sentName := fmt.Sprintf("%d送", year)
	recvName := fmt.Sprintf("%d受", year)
	mournName := fmt.Sprintf("%d喪中", year)
	for i, h := range header {
		switch h {
		case sentName:
			yc.sent = i
		case recvName:
			yc.received = i
		case mournName:
			yc.mourning = i
		}
	}
	return yc
}

func getCell(cells []string, idx int) string {
	if idx < 0 || idx >= len(cells) {
		return ""
	}
	return strings.TrimSpace(cells[idx])
}

func isChecked(val string) bool {
	return val != ""
}

func normalizePostalCode(code string) string {
	code = strings.ReplaceAll(code, "-", "")
	code = strings.ReplaceAll(code, "ー", "")
	code = strings.ReplaceAll(code, "−", "")
	code = strings.ReplaceAll(code, " ", "")
	return code
}

func parseJointNames(s string) []string {
	if s == "" {
		return nil
	}
	s = strings.ReplaceAll(s, "、", ",")
	s = strings.ReplaceAll(s, "\n", ",")
	parts := strings.Split(s, ",")
	var names []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			names = append(names, p)
		}
	}
	return names
}

func columnLetter(idx int) string {
	result := ""
	for {
		result = string(rune('A'+idx%26)) + result
		idx = idx/26 - 1
		if idx < 0 {
			break
		}
	}
	return result
}

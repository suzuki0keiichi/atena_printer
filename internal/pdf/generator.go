package pdf

import (
	"fmt"
	"unicode/utf8"

	"atena_printer/internal/config"
	"atena_printer/internal/model"

	"github.com/signintech/gopdf"
)

type Generator struct {
	pdf    *gopdf.GoPdf
	font   string
	sender config.Sender
}

func NewGenerator(fontFile string, sender config.Sender) (*Generator, error) {
	p := &gopdf.GoPdf{}
	p.Start(gopdf.Config{
		PageSize: gopdf.Rect{W: HagakiWidth, H: HagakiHeight},
		Unit:     gopdf.UnitMM,
	})

	if err := p.AddTTFFont("mincho", fontFile); err != nil {
		return nil, fmt.Errorf("フォントの読み込みに失敗: %w", err)
	}

	return &Generator{pdf: p, font: "mincho", sender: sender}, nil
}

// AddPage は1人分の宛名ページを追加する
func (g *Generator) AddPage(addr model.Address) error {
	g.pdf.AddPage()

	// 宛先郵便番号
	g.drawPostalCode(addr.PostalCode, recipientPostalX[:], recipientPostalY, recipientPostalSize)

	// 宛先住所
	g.drawVerticalText(recipientAddr1X, recipientAddrY, addr.Address1, recipientAddrSize, recipientAddrLimit)
	if addr.Address2 != "" {
		g.drawVerticalText(recipientAddr2X, recipientAddrY+5, addr.Address2, recipientAddrSize-1.5, recipientAddrLimit)
	}

	// 宛先名前
	g.drawRecipientName(addr)

	// 差出人郵便番号
	senderPostal := normalizePostal(g.sender.PostalCode)
	g.drawPostalCode(senderPostal, senderPostalX[:], senderPostalY, senderPostalSize)

	// 差出人住所
	g.drawVerticalText(senderAddr1X, senderAddrY, g.sender.Address1, senderAddrSize, senderAddrLimit)
	if g.sender.Address2 != "" {
		g.drawVerticalText(senderAddr2X, senderAddrY+2, g.sender.Address2, senderAddrSize-1, senderAddrLimit)
	}

	// 差出人名前
	senderName := g.sender.FamilyName + g.sender.GivenName
	g.drawVerticalText(senderNameX, senderNameY, senderName, senderNameSize, senderNameLimit)

	return nil
}

// Save はPDFをファイルに書き出す
func (g *Generator) Save(path string) error {
	return g.pdf.WritePdf(path)
}

func (g *Generator) drawRecipientName(addr model.Address) {
	fullName := addr.FamilyName + addr.GivenName + addr.Honorific
	nameLen := utf8.RuneCountInString(fullName)

	// 名前の長さに応じてフォントサイズを調整
	fontSize := recipientNameSize
	availableHeight := recipientNameLimit - recipientNameY
	neededHeight := float64(nameLen) * fontSize * ptToMM * 1.3
	if neededHeight > availableHeight {
		fontSize = availableHeight / (float64(nameLen) * ptToMM * 1.3)
	}

	x := recipientNameX
	// 連名がある場合は全体を少し右にずらす
	if len(addr.JointNames) > 0 {
		x += float64(len(addr.JointNames)) * jointNameSpacing / 2
	}

	// 姓の開始Y
	startY := recipientNameY

	// 名前全体をある程度中央に配置する
	totalHeight := float64(nameLen) * fontSize * ptToMM * 1.3
	if totalHeight < availableHeight {
		startY += (availableHeight - totalHeight) / 4 // 少し上寄せ
	}

	// 姓名を書く
	g.drawVerticalText(x, startY, addr.FamilyName, fontSize, recipientNameLimit)
	givenY := startY + float64(utf8.RuneCountInString(addr.FamilyName))*fontSize*ptToMM*1.3
	g.drawVerticalText(x, givenY, addr.GivenName, fontSize, recipientNameLimit)
	honorificY := givenY + float64(utf8.RuneCountInString(addr.GivenName))*fontSize*ptToMM*1.3
	g.drawVerticalText(x, honorificY, addr.Honorific, fontSize, recipientNameLimit)

	// 連名
	for i, jn := range addr.JointNames {
		jx := x - float64(i+1)*jointNameSpacing
		jNameAndHonorific := jn + addr.Honorific
		jNameLen := utf8.RuneCountInString(jNameAndHonorific)
		jFontSize := fontSize
		jNeeded := float64(jNameLen) * jFontSize * ptToMM * 1.3
		if jNeeded > (recipientNameLimit - givenY) {
			jFontSize = (recipientNameLimit - givenY) / (float64(jNameLen) * ptToMM * 1.3)
		}
		g.drawVerticalText(jx, givenY, jn, jFontSize, recipientNameLimit)
		jHonY := givenY + float64(utf8.RuneCountInString(jn))*jFontSize*ptToMM*1.3
		g.drawVerticalText(jx, jHonY, addr.Honorific, jFontSize, recipientNameLimit)
	}
}

// ptToMM はポイントをmmに変換する係数 (1pt ≈ 0.3528mm)
const ptToMM = 0.3528

func (g *Generator) drawPostalCode(code string, xs []float64, y float64, fontSize float64) {
	if err := g.pdf.SetFont(g.font, "", int(fontSize)); err != nil {
		return
	}

	runes := []rune(code)
	for i := 0; i < 7 && i < len(runes); i++ {
		ch := string(runes[i])
		w, _ := g.pdf.MeasureTextWidth(ch)
		wMM := w // gopdfはUnitMM設定済みなのでそのまま
		g.pdf.SetX(xs[i] - wMM/2)
		g.pdf.SetY(y - fontSize*ptToMM/2)
		g.pdf.Cell(nil, ch)
	}
}

func (g *Generator) drawVerticalText(x, startY float64, text string, fontSize float64, limitY float64) {
	if text == "" {
		return
	}

	// 住所の数字は全角に変換
	text = halfToFull(text)

	if err := g.pdf.SetFont(g.font, "", int(fontSize)); err != nil {
		return
	}

	charHeight := fontSize * ptToMM * 1.3 // 行送り
	y := startY

	for _, r := range text {
		if y+charHeight > limitY {
			break // 領域を超えたら打ち切り
		}

		ch := string(r)
		w, _ := g.pdf.MeasureTextWidth(ch)
		wMM := w

		dx, dy := verticalCharOffset(r, fontSize*ptToMM)

		if isVerticalRotateChar(r) {
			// 回転が必要な文字（ー、〜など）
			g.pdf.Rotate(90, x, y+charHeight/2)
			g.pdf.SetX(x - charHeight/2)
			g.pdf.SetY(y)
			g.pdf.Cell(nil, ch)
			g.pdf.RotateReset()
		} else {
			g.pdf.SetX(x - wMM/2 + dx)
			g.pdf.SetY(y + dy)
			g.pdf.Cell(nil, ch)
		}

		y += charHeight
	}
}

func normalizePostal(code string) string {
	result := make([]rune, 0, len(code))
	for _, r := range code {
		switch {
		case r >= '0' && r <= '9':
			result = append(result, r)
		case r >= '０' && r <= '９':
			result = append(result, r-'０'+'0')
		}
	}
	return string(result)
}

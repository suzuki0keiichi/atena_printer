package pdf

// はがきサイズ (mm)
const (
	HagakiWidth  = 100.0
	HagakiHeight = 148.0
)

// 宛先郵便番号の各桁の X 座標 (mm) - 日本郵便の規格に準拠
var recipientPostalX = [7]float64{
	44.8, 51.9, 59.0, // 上3桁
	67.9, 75.0, 82.1, 89.2, // 下4桁
}

const recipientPostalY = 13.5  // 郵便番号の Y 中央位置 (mm)
const recipientPostalSize = 16 // フォントサイズ (pt相当をmmスケールで調整)

// 宛先住所の開始位置
const (
	recipientAddr1X    = 83.0  // 住所1行目の X (mm)
	recipientAddr2X    = 74.0  // 住所2行目の X (mm)
	recipientAddrY     = 27.0  // 住所の開始 Y (mm)
	recipientAddrSize  = 11.0  // 住所のフォントサイズ (pt)
	recipientAddrLimit = 110.0 // 住所の下限 Y (mm)
)

// 宛先名前の位置
const (
	recipientNameX     = 56.0 // 名前の X (mm)
	recipientNameY     = 32.0 // 名前の開始 Y (mm)
	recipientNameSize  = 18.0 // 名前のフォントサイズ (pt)
	recipientNameLimit = 125.0
	jointNameSpacing   = 9.0 // 連名の列間隔 (mm)
)

// 差出人郵便番号の各桁の X 座標 (mm)
var senderPostalX = [7]float64{
	5.7, 9.6, 13.5, // 上3桁
	18.9, 22.8, 26.7, 30.6, // 下4桁
}

const senderPostalY    = 122.5 // 差出人郵便番号の Y 中央位置 (mm)
const senderPostalSize = 9     // フォントサイズ

// 差出人住所・名前の位置
const (
	senderAddr1X    = 28.0 // 差出人住所1行目の X (mm)
	senderAddr2X    = 23.5 // 差出人住所2行目の X (mm)
	senderAddrY     = 62.0 // 差出人住所の開始 Y (mm)
	senderAddrSize  = 7.5  // 差出人住所のフォントサイズ (pt)
	senderAddrLimit = 116.0
	senderNameX     = 17.0  // 差出人名前の X (mm)
	senderNameY     = 68.0  // 差出人名前の開始 Y (mm)
	senderNameSize  = 10.0  // 差出人名前のフォントサイズ (pt)
	senderNameLimit = 116.0 // 差出人名前の下限 Y (mm)
)

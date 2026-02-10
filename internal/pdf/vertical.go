package pdf

import (
	"strings"
	"unicode"
)

// halfToFull は半角数字・ハイフンを全角に変換する（縦書き用）
func halfToFull(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
			b.WriteRune(r - '0' + '０')
		case r == '-':
			b.WriteRune('ー')
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// isVerticalRotateChar は縦書き時に90度回転が必要な文字かどうかを判定する
func isVerticalRotateChar(r rune) bool {
	switch r {
	case 'ー', '〜', '～', '…', '―':
		return true
	}
	return false
}

// isSmallKana は小書き仮名かどうかを判定する
func isSmallKana(r rune) bool {
	switch r {
	case 'ぁ', 'ぃ', 'ぅ', 'ぇ', 'ぉ',
		'ゃ', 'ゅ', 'ょ', 'っ',
		'ァ', 'ィ', 'ゥ', 'ェ', 'ォ',
		'ャ', 'ュ', 'ョ', 'ッ':
		return true
	}
	return false
}

// isLatinOrDigit は半角英数字かどうかを判定する
func isLatinOrDigit(r rune) bool {
	return unicode.IsLetter(r) && r < 0x100 || unicode.IsDigit(r) && r < 0x100
}

// verticalCharOffset は縦書き時の文字ごとの位置調整を返す (dx, dy)
func verticalCharOffset(r rune, fontSize float64) (float64, float64) {
	if isSmallKana(r) {
		return fontSize * 0.1, -fontSize * 0.1
	}
	return 0, 0
}

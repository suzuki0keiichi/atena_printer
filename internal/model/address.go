package model

type Address struct {
	FamilyName string
	GivenName  string
	JointNames []string // 連名
	Honorific  string   // 敬称 (default: 様)
	PostalCode string   // ハイフンなし7桁
	Address1   string
	Address2   string
	Row        int // スプレッドシート上の行番号 (1-indexed)
}

type YearStatus struct {
	Sent     bool // 送った
	Received bool // もらった
	Mourning bool // 喪中
}

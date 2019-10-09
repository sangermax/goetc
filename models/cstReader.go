package models

var M1KEYID byte = 0x03

const (
	ReaderUnknown = 0
	DnReader1     = 1
	DnReader2     = 2
	UpReader1     = 3
	UpReader2     = 4
)

const (
	MifareS50     = 0
	MifareS70     = 1
	MifareDESFire = 2
	MifarePro     = 3
	MifareProX    = 4
)

type ReaderState struct {
	DnReaderState1 int `json:"dnReaderState1"`
	DnReaderState2 int `json:"dnReaderState2"`
	UpReaderState1 int `json:"upReaderState1"`
	UpReaderState2 int `json:"upReaderState2"`
}

type ReaderCardtype struct {
	PsamTermId string `json:"psamTermId"`
	Cardtype   int    `json:"cardtype"`
	Cardid     string `json:"cardid"`
}

type ReaderM1ReadReqData struct {
	FileID   uint16 `json:"fileID"`
	KeyID    byte   `json:"keyID"`
	FileType byte   `json:"fileType"`
	Addr     int    `json:"addr"`
	Length   int    `json:"length"`
}

type ReaderM1ReadRstData struct {
	Result int    `json:"result"`
	Data   []byte `json:"data"`
}

type ReaderM1WriteReqData struct {
	FileID   uint16 `json:"fileID"`
	KeyID    byte   `json:"keyID"`
	FileType byte   `json:"fileType"`
	Addr     int    `json:"addr"`
	Length   int    `json:"length"`
	Data     []byte `json:"data"`
}

type ReaderM1WriteRstData struct {
	Result int `json:"result"`
}

type ReaderETCReadReqData struct {
	FileID uint16 `json:"fileID"`
	Length int    `json:"length"`
}

type ReaderETCReadRstData struct {
	Result int    `json:"result"`
	Data   []byte `json:"data"`
}

//ETC卡消费
type ReaderETCPayReqData struct {
	Money   int    `json:"money"`
	Data    []byte `json:"data"`
	Paytime string `json:"paytime"`
}

type ReaderETCPayRstData struct {
	Result     int    `json:"result"`
	TradNo     []byte `json:"tradNo"`
	TermTradNo []byte `json:"termTradNo"`
	Paytime    string `json:"paytime"`
	Tac        []byte `json:"tac"`
}

//ETC余额
type ReaderETCBalanceReqData struct {
}

type ReaderETCBalanceRstData struct {
	Result  int  `json:"result"`
	Balance uint `json:"Balance"`
}

//关闭卡片
type ReaderClosecardReqData struct {
}

type ReaderClosecardRstData struct {
}

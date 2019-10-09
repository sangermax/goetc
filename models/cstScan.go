package models

type ReqScanOpenData struct {
	WorkStation string `json:"workStation"`
	ScanSpanTm  string `json:"scanSpanTm"`
}

type RstScanOpenData struct {
	ScanBar string `json:"scanBar"`
}

type ReqScanCloseData struct {
	WorkStation string `json:"workStation"`
}

type ScanStateData struct {
	UpState string `json:"upState"`
	DnState string `json:"dnState"`
}

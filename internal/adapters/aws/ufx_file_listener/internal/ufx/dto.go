package ufx

import "encoding/xml"

type Doc struct {
	XMLName        xml.Name `xml:"Doc"`
	ContractNumber string   `xml:"Originator>ContractNumber"`
	RegNumber      string   `xml:"ContractFor>Client>ClientInfo>RegNumber"`
	Parm           []Parm   `xml:"Transaction>Extra>AddData>Parm"`
	PhaseDate      string   `xml:"Billing>PhaseDate"`
	Currency       string   `xml:"Billing>Currency"`
	Amount         string   `xml:"Billing>Amount"`
}

type Parm struct {
	XMLName  xml.Name `xml:"Parm"`
	ParmCode string   `xml:"ParmCode"`
	Value    string   `xml:"Value"`
}

type FileHeader struct {
	XMLName xml.Name `xml:"FileHeader"`
	Sender  string   `xml:"Sender"`
}

type CheckSum struct {
	XMLName         xml.Name `xml:"CheckSum"`
	RecsCount       string   `xml:"RecsCount"`
	HashTotalAmount string   `xml:"HashTotalAmount"`
}

type Root struct {
	XMLName     xml.Name    `xml:"DocFile"`
	Header      FileHeader  `xml:"FileHeader"`
	DocList     DocList     `xml:"DocList"`
	FileTrailer FileTrailer `xml:"FileTrailer"`
}

type DocList struct {
	XMLName xml.Name `xml:"DocList"`
	Docs    []Doc    `xml:"Doc"`
}

type FileTrailer struct {
	XMLName  xml.Name `xml:"FileTrailer"`
	CheckSum CheckSum `xml:"CheckSum"`
}

package models

import "encoding/xml"

type (
	Ufx struct {
		XMLName     xml.Name    `xml:"DocFile"`
		Text        string      `xml:",chardata"`
		FileHeader  FileHeader  `xml:"FileHeader"`
		DocList     DocList     `xml:"DocList"`
		FileTrailer FileTrailer `xml:"FileTrailer"`
	}

	FileHeader struct {
		Text          string `xml:",chardata"`
		FileLabel     string `xml:"FileLabel"`
		FormatVersion string `xml:"FormatVersion"`
		Sender        string `xml:"Sender"`
		CreationDate  string `xml:"CreationDate"`
		CreationTime  string `xml:"CreationTime"`
		FileSeqNumber string `xml:"FileSeqNumber"`
		Receiver      string `xml:"Receiver"`
	}

	FileTrailer struct {
		Text     string   `xml:",chardata"`
		CheckSum CheckSum `xml:"CheckSum"`
	}

	CheckSum struct {
		Text            string `xml:",chardata"`
		RecsCount       string `xml:"RecsCount"`
		HashTotalAmount string `xml:"HashTotalAmount"`
	}

	DocList struct {
		Text string     `xml:",chardata"`
		Doc  []Document `xml:"Doc"`
	}

	Document struct {
		Text        string      `xml:",chardata"`
		TransType   TransType   `xml:"TransType"`
		DocRefSet   DocRefSet   `xml:"DocRefSet"`
		LocalDt     string      `xml:"LocalDt"`
		Description string      `xml:"Description"`
		ContractFor ContractFor `xml:"ContractFor"`
		Originator  Originator  `xml:"Originator"`
		Destination Destination `xml:"Destination"`
		Transaction Transaction `xml:"Transaction"`
		Billing     Billing     `xml:"Billing"`
		Status      Status      `xml:"Status"`
	}

	TransType struct {
		Text      string `xml:",chardata"`
		TransCode struct {
			Text            string `xml:",chardata"`
			MsgCode         string `xml:"MsgCode"`
			FinCategory     string `xml:"FinCategory"`
			RequestCategory string `xml:"RequestCategory"`
			ServiceClass    string `xml:"ServiceClass"`
			TransTypeCode   string `xml:"TransTypeCode"`
		} `xml:"TransCode"`
	}

	DocRefSet struct {
		Text string `xml:",chardata"`
		Parm []Parm `xml:"Parm"`
	}

	Parm struct {
		Text     string `xml:",chardata"`
		ParmCode string `xml:"ParmCode"`
		Value    string `xml:"Value"`
	}

	ContractFor struct {
		Text           string  `xml:",chardata"`
		ContractNumber string  `xml:"ContractNumber"`
		Client         Client  `xml:"Client"`
		AddData        AddData `xml:"AddData"`
	}

	Client struct {
		Text       string     `xml:",chardata"`
		ClientInfo ClientInfo `xml:"ClientInfo"`
	}

	ClientInfo struct {
		Text        string `xml:",chardata"`
		RegNumber   string `xml:"RegNumber"`
		CompanyName string `xml:"CompanyName"`
	}

	AddData struct {
		Text string      `xml:",chardata"`
		Parm AddDataParm `xml:"Parm"`
	}

	AddDataParm struct {
		Text     string `xml:",chardata"`
		ParmCode string `xml:"ParmCode"`
		Value    string `xml:"Value"`
	}

	Originator struct {
		Text           string   `xml:",chardata"`
		ContractNumber string   `xml:"ContractNumber"`
		CBSNumber      string   `xml:"CBSNumber"`
		InstInfo       InstInfo `xml:"InstInfo"`
	}

	InstInfo struct {
		Text        string `xml:",chardata"`
		Institution string `xml:"Institution"`
	}

	Destination struct {
		Text           string              `xml:",chardata"`
		ContractNumber string              `xml:"ContractNumber"`
		CBSNumber      string              `xml:"CBSNumber"`
		InstInfo       DestinationInstInfo `xml:"InstInfo"`
	}

	DestinationInstInfo struct {
		Text        string `xml:",chardata"`
		Institution string `xml:"Institution"`
		InstName    string `xml:"InstName"`
	}

	Transaction struct {
		Text  string `xml:",chardata"`
		Extra Extra  `xml:"Extra"`
	}

	Extra struct {
		Text    string       `xml:",chardata"`
		Type    string       `xml:"Type"`
		AddData ExtraAddData `xml:"AddData"`
	}

	ExtraAddData struct {
		Text string `xml:",chardata"`
		Parm []Parm `xml:"Parm"`
	}

	Billing struct {
		Text      string `xml:",chardata"`
		PhaseDate string `xml:"PhaseDate"`
		Currency  string `xml:"Currency"`
		Amount    string `xml:"Amount"`
	}

	Status struct {
		Text          string `xml:",chardata"`
		RespClass     string `xml:"RespClass"`
		RespCode      string `xml:"RespCode"`
		RespText      string `xml:"RespText"`
		PostingStatus string `xml:"PostingStatus"`
	}
)

package testhelpers

import (
	"time"

	"github.com/google/uuid"

	"github.com/saltpay/settlements-payments-system/internal/testhelpers"

	"github.com/saltpay/settlements-payments-system/black-box-tests/models"
)

type UfxBuilder struct {
	Ufx models.Ufx
}

func NewUfxBuilder() *UfxBuilder {
	docList := models.DocList{
		Text: "",
		Doc: []models.Document{
			{
				Text: "DocList",
				TransType: models.TransType{
					Text: "TransType",
					TransCode: struct {
						Text            string `xml:",chardata"`
						MsgCode         string `xml:"MsgCode"`
						FinCategory     string `xml:"FinCategory"`
						RequestCategory string `xml:"RequestCategory"`
						ServiceClass    string `xml:"ServiceClass"`
						TransTypeCode   string `xml:"TransTypeCode"`
					}{
						"Transcode",
						"RBSCRE",
						"N",
						"P",
						"T",
						"KA",
					},
				},
				DocRefSet: models.DocRefSet{
					Text: "DocRefSet",
					Parm: []models.Parm{
						{
							Text:     "Parm",
							ParmCode: "DRN",
							Value:    "22700753360",
						},
					},
				},
				LocalDt:     time.Now().Format("02-Jan-2006 15:04:05"),
				Description: "PAY_SAXO_9937507_20220627",
				ContractFor: models.ContractFor{},
				Originator: models.Originator{
					Text:           "Originator",
					ContractNumber: uuid.New().String(),
					CBSNumber:      "0-54100-SEK",
					InstInfo: struct {
						Text        string `xml:",chardata"`
						Institution string `xml:"Institution"`
					}{
						"InstInfo",
						"1501",
					},
				},
				Destination: models.Destination{
					Text:           "",
					ContractNumber: "0000-RBS_BANK_ACCOUNT",
					CBSNumber:      "0-18-11000",
					InstInfo: models.DestinationInstInfo{
						Text:        "InstInfo",
						Institution: "ERL_SWIFT",
						InstName:    "Crossborder SWIFT",
					},
				},
				Transaction: models.Transaction{
					Text: "Transaction",
					Extra: models.Extra{
						Text: "Extra",
						Type: "AddInfo",
						AddData: models.ExtraAddData{
							Text: "AddData",
							Parm: []models.Parm{
								{
									Text:     "Parm",
									ParmCode: "AC_OWNER",
									Value:    "Sideline",
								},
								{
									Text:     "Parm",
									ParmCode: "EMAIL",
									Value:    "",
								},
								{
									Text:     "Parm",
									ParmCode: "CO_CITY",
									Value:    "Reykjavík",
								},
								{
									Text:     "Parm",
									ParmCode: "CO_ADDR1",
									Value:    "Borgartúni 29",
								},
								{
									Text:     "Parm",
									ParmCode: "CO_ADDR2",
									Value:    "",
								},
								{
									Text:     "Parm",
									ParmCode: "CO_COUNTRY",
									Value:    "HUN",
								},
								{
									Text:     "Parm",
									ParmCode: "IBAN",
									Value:    "HU250301300680290006000790",
								},
								{
									Text:     "Parm",
									ParmCode: "BANK_COUNTRY",
									Value:    "HUN",
								},
								{
									Text:     "Parm",
									ParmCode: "SWIFT",
									Value:    "",
								},
								{
									Text:     "Parm",
									ParmCode: "SWIFT_REF_NUM",
									Value:    "",
								},
							},
						},
					},
				},
				Billing: models.Billing{
					Text:      "Billing",
					PhaseDate: time.Now().Format("2006-01-02"),
					Currency:  "348",
					Amount:    "1000.00",
				},
				Status: models.Status{
					Text:          "Status",
					RespClass:     "Information",
					RespCode:      "0",
					RespText:      "Successfully completed",
					PostingStatus: "Posted",
				},
			},
		},
	}

	return &UfxBuilder{
		Ufx: models.Ufx{
			Text: "s3TestUfxFile",
			FileHeader: models.FileHeader{
				Text:          "FileHeader",
				FileLabel:     "ORDER",
				FormatVersion: "2.2",
				Sender:        "SAXO_FAKE",
				CreationDate:  time.Now().Format("2006-01-02"),
				CreationTime:  time.Now().Format("15:04:05"),
				FileSeqNumber: testhelpers.RandomString(),
				Receiver:      "BORGUN_FAKE",
			},
			DocList: docList,
			FileTrailer: models.FileTrailer{
				Text: "FileTrailer",
				CheckSum: models.CheckSum{
					Text:            "CheckSum",
					RecsCount:       "1",
					HashTotalAmount: "1000.00",
				},
			},
		},
	}
}

func (u *UfxBuilder) Build() models.Ufx {
	return u.Ufx
}

func (u *UfxBuilder) WithMerchantID(mid string) *UfxBuilder {
	u.Ufx.DocList.Doc[0].Originator.ContractNumber = mid
	return u
}

func (u *UfxBuilder) WithRegNumber(regNumber string) *UfxBuilder {
	u.Ufx.DocList.Doc[0].ContractFor.Client.ClientInfo.RegNumber = regNumber
	return u
}

func (u *UfxBuilder) WithIBAN(iban string) *UfxBuilder {
	parmList := u.Ufx.DocList.Doc[0].Transaction.Extra.AddData.Parm
	for i, par := range parmList {
		if par.ParmCode == "IBAN" {
			parmList[i].Value = iban
		}
	}
	return u
}

func (u *UfxBuilder) WithSWIFT(swift string) *UfxBuilder {
	parmList := u.Ufx.DocList.Doc[0].Transaction.Extra.AddData.Parm
	for i, par := range parmList {
		if par.ParmCode == "SWIFT" {
			parmList[i].Value = swift
		}
	}
	return u
}

func (u *UfxBuilder) WithSender(sender string) *UfxBuilder {
	u.Ufx.FileHeader.Sender = sender

	return u
}

func (u *UfxBuilder) WithCurrency(currencyCode models.CurrencyCode) *UfxBuilder {
	convertedCurrency := models.CurrenciesToIso[currencyCode]
	u.Ufx.DocList.Doc[0].Billing.Currency = convertedCurrency

	return u
}

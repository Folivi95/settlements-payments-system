package ufx_file_listener

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	zapctx "github.com/saltpay/go-zap-ctx"
	"go.uber.org/zap"

	"github.com/saltpay/settlements-payments-system/internal/adapters/aws/ufx_file_listener/internal/ufx"
	"github.com/saltpay/settlements-payments-system/internal/adapters/aws/ufx_file_listener/internal/utils"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

type UfxToPaymentInstructionsConverter struct{}

func NewUfxToPaymentInstructionsConverter() UfxToPaymentInstructionsConverter {
	return UfxToPaymentInstructionsConverter{}
}

func (ufxConverter *UfxToPaymentInstructionsConverter) ConvertUfx(ctx context.Context, ufxFileContents io.Reader, ufxFileName string) (models.IncomingInstructions, error) {
	decoder := xml.NewDecoder(ufxFileContents)
	var fileHeader ufx.FileHeader
	var docs []ufx.Doc
	for {
		token, err := decoder.Token()

		// Read until end of file
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error parsing the UFX XML: %s", err.Error())
		}
		switch ty := token.(type) {
		case xml.StartElement:
			if ty.Name.Local == "FileHeader" {
				err = decoder.DecodeElement(&fileHeader, &ty)
				if err != nil {
					return nil, fmt.Errorf("error parsing the UFX XML: %s", err.Error())
				}
			}
			if ty.Name.Local == "Doc" {
				var doc ufx.Doc
				err = decoder.DecodeElement(&doc, &ty)
				if err != nil {
					return nil, fmt.Errorf("error parsing the UFX XML: %s", err.Error())
				}
				docs = append(docs, doc)
			}
		}
	}

	return ufxConverter.createIncomingInstructions(ctx, docs, ufxFileName, fileHeader)
}

func (ufxConverter *UfxToPaymentInstructionsConverter) FilterCurrency(ufxFileContents io.Reader, currency models.CurrencyCode) ([]byte, error) {
	decoder := xml.NewDecoder(ufxFileContents)
	var fileHeader ufx.FileHeader
	var docs []ufx.Doc
	var root ufx.Root
	for {
		token, err := decoder.Token()

		// Read until end of file
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error parsing the UFX XML: %s", err.Error())
		}
		switch ty := token.(type) {
		case xml.StartElement:
			if ty.Name.Local == "FileHeader" {
				err = decoder.DecodeElement(&fileHeader, &ty)
				if err != nil {
					return nil, fmt.Errorf("error parsing the UFX XML: %s", err.Error())
				}
			}
			if ty.Name.Local == "Doc" {
				var doc ufx.Doc
				err = decoder.DecodeElement(&doc, &ty)
				if err != nil {
					return nil, fmt.Errorf("error parsing the UFX XML: %s", err.Error())
				}
				convertedCurrency := models.CurrenciesToIso[currency]
				if doc.Currency == convertedCurrency {
					docs = append(docs, doc)
				}
			}
		}
	}

	root.Header = fileHeader
	root.DocList.Docs = append(root.DocList.Docs, docs...)

	filteredXML, err := xml.MarshalIndent(root, "  ", "    ")
	if err != nil {
		return nil, fmt.Errorf("error converting the file to xml: %s", err.Error())
	}

	return filteredXML, nil
}

func (ufxConverter *UfxToPaymentInstructionsConverter) createIncomingInstructions(ctx context.Context, docs []ufx.Doc, fileName string, fileHeader ufx.FileHeader) ([]models.IncomingInstruction, error) {
	incomingInstructions := make([]models.IncomingInstruction, len(docs))
	for index, doc := range docs {
		request, err := ufxConverter.createPaymentInstruction(ctx, doc, fileName, fileHeader)
		if err != nil { // techdebt-high: when one payment instruction can't be created, skip and log it, and process the others
			return nil, err
		}
		incomingInstructions[index] = request
	}

	sort.Sort(models.ByCurrencyPriority(incomingInstructions))

	return incomingInstructions, nil
}

func (ufxConverter *UfxToPaymentInstructionsConverter) createPaymentInstruction(ctx context.Context, doc ufx.Doc, fileName string, fileHeader ufx.FileHeader) (models.IncomingInstruction, error) {
	executionDate, err := utils.ConvertToDate(doc.PhaseDate)
	if err != nil {
		return models.IncomingInstruction{}, err
	}

	instruction := newIncomingInstructionFromUFX(doc, fileName, executionDate, fileHeader)

	if models.IsISLSender(fileHeader.Sender) {
		ufxConverter.handleRBSender(ctx, &instruction)
	}

	if models.IsSaxoSender(fileHeader.Sender) {
		ufxConverter.handleSaxoSender(&instruction)
	}

	return instruction, nil
}

func (ufxConverter *UfxToPaymentInstructionsConverter) handleRBSender(ctx context.Context, instruction *models.IncomingInstruction) {
	iban := instruction.Merchant.Account.AccountNumber

	if len(iban) > 0 {
		return
	}

	// generate iban using kennitala & account number
	kennitala := instruction.Merchant.RegNumber
	accountNumber := instruction.Merchant.Account.Swift
	iban, err := utils.GenerateIcelandicIBAN(kennitala, accountNumber)
	if err != nil {
		zapctx.Error(ctx, "failed to generate the IBAN from kennitala and account number",
			zap.Any("instruction", instruction),
			zap.Error(err),
		)
		return
	}
	instruction.Merchant.Account.AccountNumber = iban
}

func (ufxConverter *UfxToPaymentInstructionsConverter) handleSaxoSender(instruction *models.IncomingInstruction) {
	instruction.Merchant.Account.Country = utils.ExtractCountryCode(instruction.Merchant.Account.Swift)
}

func newIncomingInstructionFromUFX(doc ufx.Doc, fileName string, executionDate time.Time, fileHeader ufx.FileHeader) models.IncomingInstruction {
	return models.IncomingInstruction{
		Merchant: models.Merchant{
			ContractNumber: doc.ContractNumber,
			RegNumber:      doc.RegNumber,
			Name:           utils.GetParamValue(doc.Parm, "AC_OWNER"),
			Email:          utils.GetParamValue(doc.Parm, "EMAIL"),
			Address: models.Address{
				City:         utils.GetParamValue(doc.Parm, "CO_CITY"),
				AddressLine1: utils.GetParamValue(doc.Parm, "CO_ADDR1"),
				AddressLine2: utils.GetParamValue(doc.Parm, "CO_ADDR2"),
				Country:      utils.GetParamValue(doc.Parm, "CO_COUNTRY"),
			},
			Account: models.Account{
				AccountNumber:        utils.GetParamValue(doc.Parm, "IBAN"),
				Swift:                utils.GetParamValue(doc.Parm, "SWIFT"),
				SwiftReferenceNumber: utils.GetParamValue(doc.Parm, "SWIFT_REF_NUM"),
				BankCountry:          utils.GetParamValue(doc.Parm, "BANK_COUNTRY"),
			},
			HighRisk: strings.Contains(fileName, "_HR_"),
		},
		Payment: models.Payment{
			Sender: models.Sender{},
			Amount: doc.Amount,
			Currency: models.Currency{
				IsoNumber: doc.Currency,
				IsoCode:   models.GetCurrencyIsoCode(doc.Currency),
			},
			ExecutionDate: executionDate,
		},
		Metadata: models.Metadata{
			Source:   "Way4",
			FileType: "UFX",
			Filename: fileName,
			Sender:   fileHeader.Sender,
		},
	}
}

package testhelpers

import (
	"time"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

type UfxAndIncomingInstruction struct {
	UfxFileName         string
	UfxFileContents     string
	IncomingInstruction models.IncomingInstruction
}

func ValidUfxAndIncomingInstruction() UfxAndIncomingInstruction {
	return UfxAndIncomingInstruction{
		UfxFileName:     "test.ufx",
		UfxFileContents: testUfx(),
		IncomingInstruction: models.IncomingInstruction{
			Merchant: models.Merchant{
				ContractNumber: "9000000",
				RegNumber:      "4000000000",
				Name:           "Ms Big Shot Merchant",
				Email:          "testemail@testmerchant.is",
				Address: models.Address{
					Country:      "GBR",
					City:         "TestCity",
					AddressLine1: "Testaddress 9",
					AddressLine2: "240 TestCity",
				},
				Account: models.Account{
					AccountNumber:        "GB33BUKB20201555555555",
					Swift:                "050026000000",
					Country:              "26",
					SwiftReferenceNumber: "",
					BankCountry:          "ISL",
				},
				HighRisk: false,
			},
			Metadata: models.Metadata{Source: "Way4", Filename: "test.ufx", FileType: "UFX", Sender: "SAXO"},
			Payment: models.Payment{
				Sender:        models.Sender{},
				Amount:        "10",
				Currency:      models.Currency{IsoCode: "EUR", IsoNumber: "978"},
				ExecutionDate: time.Date(2021, time.June, 30, 0, 0, 0, 0, time.UTC),
			},
		},
	}
}

func testUfx() string {
	return `<?xml version="1.0" encoding="utf-8"?>
<DocFile>
	<FileHeader>
		<FileLabel>ORDER</FileLabel>
		<FormatVersion>2.2</FormatVersion>
		<Sender>SAXO</Sender>
		<CreationDate>2021-02-04</CreationDate>
		<CreationTime>07:34:07</CreationTime>
		<FileSeqNumber>6</FileSeqNumber>
		<Receiver>BORGUN</Receiver>
	</FileHeader>
	<DocList>
		<Doc>
			<TransType>
				<TransCode>
					<MsgCode>RBSCRE</MsgCode>
					<FinCategory>N</FinCategory>
					<RequestCategory>P</RequestCategory>
					<ServiceClass>T</ServiceClass>
					<TransTypeCode>KA</TransTypeCode>
				</TransCode>
			</TransType>
			<DocRefSet>
				<Parm>
					<ParmCode>DRN</ParmCode>
					<Value>13907938650</Value>
				</Parm>
				<Parm>
					<ParmCode>SRN</ParmCode>
					<Value>SO-000000013907938650</Value>
				</Parm>
			</DocRefSet>
			<LocalDt>2021-06-30 05:42:07</LocalDt>
			<Description>PAY_RB_9778256_20210204</Description>
			<ContractFor>
				<ContractNumber>0</ContractNumber>
				<Client>
					<ClientInfo>
						<RegNumber>4000000000</RegNumber>
						<CompanyName>Test merchant</CompanyName>
					</ClientInfo>
				</Client>
			</ContractFor>
			<Originator>
				<ContractNumber>9000000</ContractNumber>
				<CBSNumber>0-54100-ISK</CBSNumber>
				<InstInfo>
					<Institution>1501</Institution>
				</InstInfo>
			</Originator>
			<Destination>
				<ContractNumber>1501-RBS_BANK_ACCOUNT</ContractNumber>
				<CBSNumber>0-15-3555</CBSNumber>
				<InstInfo>
					<Institution>INL_RB</Institution>
					<InstName>Domestic RB</InstName>
				</InstInfo>
			</Destination>
			<Transaction>
				<Extra>
					<Type>AddInfo</Type>
					<AddData>
						<Parm>
							<ParmCode>BORGUN_REGION</ParmCode>
							<Value/>
						</Parm>
						<Parm>
							<ParmCode>trade_country</ParmCode>
							<Value>ISL</Value>
						</Parm>
						<Parm>
							<ParmCode>BANK_CITY</ParmCode>
							<Value>Reykjanesbar</Value>
						</Parm>
						<Parm>
							<ParmCode>trade_state</ParmCode>
							<Value/>
						</Parm>
						<Parm>
							<ParmCode>CO_ADDR3</ParmCode>
							<Value/>
						</Parm>
						<Parm>
							<ParmCode>CO_ADDR2</ParmCode>
							<Value>240 TestCity</Value>
						</Parm>
						<Parm>
							<ParmCode>CO_ADDR1</ParmCode>
							<Value>Testaddress 9</Value>
						</Parm>
						<Parm>
							<ParmCode>BY_BATCH</ParmCode>
							<Value/>
						</Parm>
						<Parm>
							<ParmCode>AC_OWNER_COUNTRY</ParmCode>
							<Value/>
						</Parm>
						<Parm>
							<ParmCode>SWIFT</ParmCode>
							<Value>050026000000</Value>
						</Parm>
						<Parm>
							<ParmCode>BATCH_ID</ParmCode>
							<Value>STND</Value>
						</Parm>
						<Parm>
							<ParmCode>EMAIL</ParmCode>
							<Value>testemail@testmerchant.is</Value>
						</Parm>
						<Parm>
							<ParmCode>ORDER_MIN_PARM</ParmCode>
							<Value>MIN_PAY_AMNT_ISK</Value>
						</Parm>
						<Parm>
							<ParmCode>IS_REVERSAL</ParmCode>
							<Value>N</Value>
						</Parm>
						<Parm>
							<ParmCode>BANK_COUNTRY</ParmCode>
							<Value>ISL</Value>
						</Parm>
						<Parm>
							<ParmCode>CO_NAME</ParmCode>
							<Value>Test merchant</Value>
						</Parm>
						<Parm>
							<ParmCode>POSTING_DATE</ParmCode>
							<Value>CURRENT</Value>
						</Parm>
						<Parm>
							<ParmCode>BANK_STATE_CODE</ParmCode>
							<Value/>
						</Parm>
						<Parm>
							<ParmCode>BANK_EMAIL</ParmCode>
							<Value/>
						</Parm>
						<Parm>
							<ParmCode>ISO_CONTRACT</ParmCode>
							<Value>N</Value>
						</Parm>
						<Parm>
							<ParmCode>USE_DOC_DATA</ParmCode>
							<Value>Y</Value>
						</Parm>
						<Parm>
							<ParmCode>IF_CS_VALUE</ParmCode>
							<Value>SPLIT</Value>
						</Parm>
						<Parm>
							<ParmCode>ORDER_CURRENCY</ParmCode>
							<Value>ISK</Value>
						</Parm>
						<Parm>
							<ParmCode>R_ISK</ParmCode>
							<Value>0</Value>
						</Parm>
						<Parm>
							<ParmCode>BANK_ADDR</ParmCode>
							<Value/>
						</Parm>
						<Parm>
							<ParmCode>IN_WRK_DAY</ParmCode>
							<Value/>
						</Parm>
						<Parm>
							<ParmCode>SWIFT_REF_NUM</ParmCode>  
							<Value/>
						</Parm>
						<Parm>
							<ParmCode>trade_city</ParmCode>
							<Value>TestCity</Value>
						</Parm>
						<Parm>
							<ParmCode>trade_zip</ParmCode>
							<Value>241</Value>
						</Parm>
						<Parm>
							<ParmCode>BANK_POST_CODE</ParmCode>
							<Value>230</Value>
						</Parm>
						<Parm>
							<ParmCode>IBAN</ParmCode>
							<Value>GB33BUKB20201555555555</Value>
						</Parm>
						<Parm>
							<ParmCode>ADDRESS_TYPE</ParmCode>
							<Value>BNK_ADDR_ISK</Value>
						</Parm>
						<Parm>
							<ParmCode>AC_OWNER_ADDR</ParmCode>
							<Value/>
						</Parm>
						<Parm>
							<ParmCode>PAYMENT_METHOD</ParmCode>
							<Value>RB</Value>
						</Parm>
						<Parm>
							<ParmCode>CO_COUNTRY</ParmCode>
							<Value>GBR</Value>
						</Parm>
						<Parm>
							<ParmCode>AC_OWNER_CITY</ParmCode>
							<Value/>
						</Parm>
						<Parm>
							<ParmCode>PAYMENT_TYPE</ParmCode>
							<Value>RB</Value>
						</Parm>
						<Parm>
							<ParmCode>BANK_AC_GL</ParmCode>
							<Value/>
						</Parm>
						<Parm>
							<ParmCode>ORDER_ID</ParmCode>
							<Value>89000000</Value>
						</Parm>
						<Parm>
							<ParmCode>CO_CITY</ParmCode>
							<Value>TestCity</Value>
						</Parm>
						<Parm>
							<ParmCode>IF_CS_TYPE</ParmCode>
							<Value>SPLITTED_PAYMENTS</Value>
						</Parm>
						<Parm>
							<ParmCode>AC_OWNER</ParmCode>
							<Value>Ms Big Shot Merchant</Value>
						</Parm>
						<Parm>
							<ParmCode>trade_addr_1</ParmCode>
							<Value>Test merchant</Value>
						</Parm>
						<Parm>
							<ParmCode>CUSTOM_AMOUNT</ParmCode>
							<Value/>
						</Parm>
						<Parm>
							<ParmCode>trade_addr_2</ParmCode>
							<Value>Testaddress 9</Value>
						</Parm>
						<Parm>
							<ParmCode>BILLING_DAY</ParmCode>
							<Value>Y</Value>
						</Parm>
						<Parm>
							<ParmCode>BANK_NAME</ParmCode>
							<Value>Íslandsbanki</Value>
						</Parm>
					</AddData>
				</Extra>
			</Transaction>
			<Billing>
				<PhaseDate>2021-06-30</PhaseDate>
				<Currency>978</Currency>
				<Amount>10</Amount>
			</Billing>
			<Status>
				<RespClass>Information</RespClass>
				<RespCode>0</RespCode>
				<RespText>Successfully completed</RespText>
				<PostingStatus>Posted</PostingStatus>
			</Status>
		</Doc>
	</DocList>
	<FileTrailer>
		<CheckSum>
			<RecsCount>1</RecsCount>
			<HashTotalAmount>9668</HashTotalAmount>
		</CheckSum>
	</FileTrailer>
</DocFile>`
}

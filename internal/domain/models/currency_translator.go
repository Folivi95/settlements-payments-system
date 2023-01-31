package models

import "fmt"

func GetCurrencyIsoCode(currencyIsoNumber string) CurrencyCode {
	padded := zeroPadNumber(currencyIsoNumber)
	if code, exists := currencies[padded]; exists {
		return code
	}
	return ""
}

func zeroPadNumber(number string) string {
	return fmt.Sprintf("%03s", number)
}

// GetAllCurrencyCodes builds a slice with all the currencies registered in the system.
func GetAllCurrencyCodes() []CurrencyCode {
	codes := make([]CurrencyCode, 0, len(currencies))
	for _, v := range currencies {
		codes = append(codes, v)
	}
	return codes
}

// ISO 4217 currency codes.
var currencies = map[string]CurrencyCode{
	"008": ALL,
	"012": DZD,
	"032": ARS,
	"036": AUD,
	"044": BSD,
	"048": BHD,
	"050": BDT,
	"051": AMD,
	"052": BBD,
	"060": BMD,
	"064": BTN,
	"068": BOB,
	"072": BWP,
	"084": BZD,
	"090": SBD,
	"096": BND,
	"104": MMK,
	"108": BIF,
	"116": KHR,
	"124": CAD,
	"132": CVE,
	"136": KYD,
	"144": LKR,
	"152": CLP,
	"156": CNY,
	"170": COP,
	"174": KMF,
	"188": CRC,
	"191": HRK,
	"192": CUP,
	"203": CZK,
	"208": DKK,
	"214": DOP,
	"222": SVC,
	"230": ETB,
	"232": ERN,
	"238": FKP,
	"242": FJD,
	"262": DJF,
	"270": GMD,
	"292": GIP,
	"320": GTQ,
	"324": GNF,
	"328": GYD,
	"332": HTG,
	"340": HNL,
	"344": HKD,
	"348": HUF,
	"352": ISK,
	"356": INR,
	"360": IDR,
	"364": IRR,
	"368": IQD,
	"376": ILS,
	"388": JMD,
	"392": JPY,
	"398": KZT,
	"400": JOD,
	"404": KES,
	"408": KPW,
	"410": KRW,
	"414": KWD,
	"417": KGS,
	"418": LAK,
	"422": LBP,
	"426": LSL,
	"430": LRD,
	"434": LYD,
	"446": MOP,
	"454": MWK,
	"458": MYR,
	"462": MVR,
	"478": MRO,
	"480": MUR,
	"484": MXN,
	"496": MNT,
	"498": MDL,
	"504": MAD,
	"512": OMR,
	"516": NAD,
	"524": NPR,
	"532": ANG,
	"533": AWG,
	"548": VUV,
	"554": NZD,
	"558": NIO,
	"566": NGN,
	"578": NOK,
	"586": PKR,
	"590": PAB,
	"598": PGK,
	"600": PYG,
	"604": PEN,
	"608": PHP,
	"634": QAR,
	"643": RUB,
	"646": RWF,
	"654": SHP,
	"678": STD,
	"682": SAR,
	"690": SCR,
	"694": SLL,
	"702": SGD,
	"704": VND,
	"706": SOS,
	"710": ZAR,
	"728": SSP,
	"748": SZL,
	"752": SEK,
	"756": CHF,
	"760": SYP,
	"764": THB,
	"776": TOP,
	"780": TTD,
	"784": AED,
	"788": TND,
	"800": UGX,
	"807": MKD,
	"818": EGP,
	"826": GBP,
	"834": TZS,
	"840": USD,
	"858": UYU,
	"860": UZS,
	"882": WST,
	"886": YER,
	"901": TWD,
	"928": VES,
	"931": CUC,
	"932": ZWL,
	"933": BYN,
	"934": TMT,
	"936": GHS,
	"937": VEF,
	"938": SDG,
	"940": UYI,
	"941": RSD,
	"943": MZN,
	"944": AZN,
	"946": RON,
	"947": CHE,
	"948": CHW,
	"949": TRY,
	"950": XAF,
	"951": XCD,
	"952": XOF,
	"953": XPF,
	"967": ZMW,
	"968": SRD,
	"969": MGA,
	"970": COU,
	"971": AFN,
	"972": TJS,
	"973": AOA,
	"974": BYR,
	"975": BGN,
	"976": CDF,
	"977": BAM,
	"978": EUR,
	"979": MXV,
	"980": UAH,
	"981": GEL,
	"984": BOV,
	"985": PLN,
	"986": BRL,
	"990": CLF,
	"997": USN,
}

// CurrenciesToIso Map currency code to ISO 4217.
var CurrenciesToIso = map[CurrencyCode]string{
	ALL: "008",
	DZD: "012",
	ARS: "032",
	AUD: "036",
	BSD: "044",
	BHD: "048",
	BDT: "050",
	AMD: "051",
	BBD: "052",
	BMD: "060",
	BTN: "064",
	BOB: "068",
	BWP: "072",
	BZD: "084",
	SBD: "090",
	BND: "096",
	MMK: "104",
	BIF: "108",
	KHR: "116",
	CAD: "124",
	CVE: "132",
	KYD: "136",
	LKR: "144",
	CLP: "152",
	CNY: "156",
	COP: "170",
	KMF: "174",
	CRC: "188",
	HRK: "191",
	CUP: "192",
	CZK: "203",
	DKK: "208",
	DOP: "214",
	SVC: "222",
	ETB: "230",
	ERN: "232",
	FKP: "238",
	FJD: "242",
	DJF: "262",
	GMD: "270",
	GIP: "292",
	GTQ: "320",
	GNF: "324",
	GYD: "328",
	HTG: "332",
	HNL: "340",
	HKD: "344",
	HUF: "348",
	ISK: "352",
	INR: "356",
	IDR: "360",
	IRR: "364",
	IQD: "368",
	ILS: "376",
	JMD: "388",
	JPY: "392",
	KZT: "398",
	JOD: "400",
	KES: "404",
	KPW: "408",
	KRW: "410",
	KWD: "414",
	KGS: "417",
	LAK: "418",
	LBP: "422",
	LSL: "426",
	LRD: "430",
	LYD: "434",
	MOP: "446",
	MWK: "454",
	MYR: "458",
	MVR: "462",
	MRO: "478",
	MUR: "480",
	MXN: "484",
	MNT: "496",
	MDL: "498",
	MAD: "504",
	OMR: "512",
	NAD: "516",
	NPR: "524",
	ANG: "532",
	AWG: "533",
	VUV: "548",
	NZD: "554",
	NIO: "558",
	NGN: "566",
	NOK: "578",
	PKR: "586",
	PAB: "590",
	PGK: "598",
	PYG: "600",
	PEN: "604",
	PHP: "608",
	QAR: "634",
	RUB: "643",
	RWF: "646",
	SHP: "654",
	STD: "678",
	SAR: "682",
	SCR: "690",
	SLL: "694",
	SGD: "702",
	VND: "704",
	SOS: "706",
	ZAR: "710",
	SSP: "728",
	SZL: "748",
	SEK: "752",
	CHF: "756",
	SYP: "760",
	THB: "764",
	TOP: "776",
	TTD: "780",
	AED: "784",
	TND: "788",
	UGX: "800",
	MKD: "807",
	EGP: "818",
	GBP: "826",
	TZS: "834",
	USD: "840",
	UYU: "858",
	UZS: "860",
	WST: "882",
	YER: "886",
	TWD: "901",
	VES: "928",
	CUC: "931",
	ZWL: "932",
	BYN: "933",
	TMT: "934",
	GHS: "936",
	VEF: "937",
	SDG: "938",
	UYI: "940",
	RSD: "941",
	MZN: "943",
	AZN: "944",
	RON: "946",
	CHE: "947",
	CHW: "948",
	TRY: "949",
	XAF: "950",
	XCD: "951",
	XOF: "952",
	XPF: "953",
	ZMW: "967",
	SRD: "968",
	MGA: "969",
	COU: "970",
	AFN: "971",
	TJS: "972",
	AOA: "973",
	BYR: "974",
	BGN: "975",
	CDF: "976",
	BAM: "977",
	EUR: "978",
	MXV: "979",
	UAH: "980",
	GEL: "981",
	BOV: "984",
	PLN: "985",
	BRL: "986",
	CLF: "990",
	USN: "997",
}

type CurrencyCode string

const (
	ALL CurrencyCode = "ALL"
	DZD CurrencyCode = "DZD"
	ARS CurrencyCode = "ARS"
	AUD CurrencyCode = "AUD"
	BSD CurrencyCode = "BSD"
	BHD CurrencyCode = "BHD"
	BDT CurrencyCode = "BDT"
	AMD CurrencyCode = "AMD"
	BBD CurrencyCode = "BBD"
	BMD CurrencyCode = "BMD"
	BTN CurrencyCode = "BTN"
	BOB CurrencyCode = "BOB"
	BWP CurrencyCode = "BWP"
	BZD CurrencyCode = "BZD"
	SBD CurrencyCode = "SBD"
	BND CurrencyCode = "BND"
	MMK CurrencyCode = "MMK"
	BIF CurrencyCode = "BIF"
	KHR CurrencyCode = "KHR"
	CAD CurrencyCode = "CAD"
	CVE CurrencyCode = "CVE"
	KYD CurrencyCode = "KYD"
	LKR CurrencyCode = "LKR"
	CLP CurrencyCode = "CLP"
	CNY CurrencyCode = "CNY"
	COP CurrencyCode = "COP"
	KMF CurrencyCode = "KMF"
	CRC CurrencyCode = "CRC"
	HRK CurrencyCode = "HRK"
	CUP CurrencyCode = "CUP"
	CZK CurrencyCode = "CZK"
	DKK CurrencyCode = "DKK"
	DOP CurrencyCode = "DOP"
	SVC CurrencyCode = "SVC"
	ETB CurrencyCode = "ETB"
	ERN CurrencyCode = "ERN"
	FKP CurrencyCode = "FKP"
	FJD CurrencyCode = "FJD"
	DJF CurrencyCode = "DJF"
	GMD CurrencyCode = "GMD"
	GIP CurrencyCode = "GIP"
	GTQ CurrencyCode = "GTQ"
	GNF CurrencyCode = "GNF"
	GYD CurrencyCode = "GYD"
	HTG CurrencyCode = "HTG"
	HNL CurrencyCode = "HNL"
	HKD CurrencyCode = "HKD"
	HUF CurrencyCode = "HUF"
	ISK CurrencyCode = "ISK"
	INR CurrencyCode = "INR"
	IDR CurrencyCode = "IDR"
	IRR CurrencyCode = "IRR"
	IQD CurrencyCode = "IQD"
	ILS CurrencyCode = "ILS"
	JMD CurrencyCode = "JMD"
	JPY CurrencyCode = "JPY"
	KZT CurrencyCode = "KZT"
	JOD CurrencyCode = "JOD"
	KES CurrencyCode = "KES"
	KPW CurrencyCode = "KPW"
	KRW CurrencyCode = "KRW"
	KWD CurrencyCode = "KWD"
	KGS CurrencyCode = "KGS"
	LAK CurrencyCode = "LAK"
	LBP CurrencyCode = "LBP"
	LSL CurrencyCode = "LSL"
	LRD CurrencyCode = "LRD"
	LYD CurrencyCode = "LYD"
	MOP CurrencyCode = "MOP"
	MWK CurrencyCode = "MWK"
	MYR CurrencyCode = "MYR"
	MVR CurrencyCode = "MVR"
	MRO CurrencyCode = "MRO"
	MUR CurrencyCode = "MUR"
	MXN CurrencyCode = "MXN"
	MNT CurrencyCode = "MNT"
	MDL CurrencyCode = "MDL"
	MAD CurrencyCode = "MAD"
	OMR CurrencyCode = "OMR"
	NAD CurrencyCode = "NAD"
	NPR CurrencyCode = "NPR"
	ANG CurrencyCode = "ANG"
	AWG CurrencyCode = "AWG"
	VUV CurrencyCode = "VUV"
	NZD CurrencyCode = "NZD"
	NIO CurrencyCode = "NIO"
	NGN CurrencyCode = "NGN"
	NOK CurrencyCode = "NOK"
	PKR CurrencyCode = "PKR"
	PAB CurrencyCode = "PAB"
	PGK CurrencyCode = "PGK"
	PYG CurrencyCode = "PYG"
	PEN CurrencyCode = "PEN"
	PHP CurrencyCode = "PHP"
	QAR CurrencyCode = "QAR"
	RUB CurrencyCode = "RUB"
	RWF CurrencyCode = "RWF"
	SHP CurrencyCode = "SHP"
	STD CurrencyCode = "STD"
	SAR CurrencyCode = "SAR"
	SCR CurrencyCode = "SCR"
	SLL CurrencyCode = "SLL"
	SGD CurrencyCode = "SGD"
	VND CurrencyCode = "VND"
	SOS CurrencyCode = "SOS"
	ZAR CurrencyCode = "ZAR"
	SSP CurrencyCode = "SSP"
	SZL CurrencyCode = "SZL"
	SEK CurrencyCode = "SEK"
	CHF CurrencyCode = "CHF"
	SYP CurrencyCode = "SYP"
	THB CurrencyCode = "THB"
	TOP CurrencyCode = "TOP"
	TTD CurrencyCode = "TTD"
	AED CurrencyCode = "AED"
	TND CurrencyCode = "TND"
	UGX CurrencyCode = "UGX"
	MKD CurrencyCode = "MKD"
	EGP CurrencyCode = "EGP"
	GBP CurrencyCode = "GBP"
	TZS CurrencyCode = "TZS"
	USD CurrencyCode = "USD"
	UYU CurrencyCode = "UYU"
	UZS CurrencyCode = "UZS"
	WST CurrencyCode = "WST"
	YER CurrencyCode = "YER"
	TWD CurrencyCode = "TWD"
	VES CurrencyCode = "VES"
	CUC CurrencyCode = "CUC"
	ZWL CurrencyCode = "ZWL"
	BYN CurrencyCode = "BYN"
	TMT CurrencyCode = "TMT"
	GHS CurrencyCode = "GHS"
	VEF CurrencyCode = "VEF"
	SDG CurrencyCode = "SDG"
	UYI CurrencyCode = "UYI"
	RSD CurrencyCode = "RSD"
	MZN CurrencyCode = "MZN"
	AZN CurrencyCode = "AZN"
	RON CurrencyCode = "RON"
	CHE CurrencyCode = "CHE"
	CHW CurrencyCode = "CHW"
	TRY CurrencyCode = "TRY"
	XAF CurrencyCode = "XAF"
	XCD CurrencyCode = "XCD"
	XOF CurrencyCode = "XOF"
	XPF CurrencyCode = "XPF"
	ZMW CurrencyCode = "ZMW"
	SRD CurrencyCode = "SRD"
	MGA CurrencyCode = "MGA"
	COU CurrencyCode = "COU"
	AFN CurrencyCode = "AFN"
	TJS CurrencyCode = "TJS"
	AOA CurrencyCode = "AOA"
	BYR CurrencyCode = "BYR"
	BGN CurrencyCode = "BGN"
	CDF CurrencyCode = "CDF"
	BAM CurrencyCode = "BAM"
	EUR CurrencyCode = "EUR"
	MXV CurrencyCode = "MXV"
	UAH CurrencyCode = "UAH"
	GEL CurrencyCode = "GEL"
	BOV CurrencyCode = "BOV"
	PLN CurrencyCode = "PLN"
	BRL CurrencyCode = "BRL"
	CLF CurrencyCode = "CLF"
	USN CurrencyCode = "USN"
)

package utils

import "github.com/saltpay/settlements-payments-system/internal/adapters/aws/ufx_file_listener/internal/ufx"

func GetParamValue(parms []ufx.Parm, identifier string) string {
	for _, parm := range parms {
		if parm.ParmCode == identifier {
			return parm.Value
		}
	}
	return ""
}

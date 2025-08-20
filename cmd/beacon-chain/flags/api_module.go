package flags

import "strings"

const QrysmAPIModule string = "qrysm"
const QRLAPIModule string = "qrl"

func EnableHTTPQrysmAPI(httpModules string) bool {
	return enableAPI(httpModules, QrysmAPIModule)
}

func EnableHTTPQRLAPI(httpModules string) bool {
	return enableAPI(httpModules, QRLAPIModule)
}

func enableAPI(httpModules, api string) bool {
	for _, m := range strings.Split(httpModules, ",") {
		if strings.EqualFold(m, api) {
			return true
		}
	}
	return false
}

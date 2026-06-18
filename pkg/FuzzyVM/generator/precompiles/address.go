package precompiles

import "github.com/theQRL/go-qrl/common"

func mustAddress(s string) common.Address {
	addr, err := common.NewAddressFromString(s)
	if err != nil {
		panic(err) // lint:nopanic
	}
	return addr
}

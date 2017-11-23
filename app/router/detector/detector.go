package detector

import (
	"github.com/edmocosta/router-config-client/app/router/intelbras"
	"fmt"
	"github.com/edmocosta/router-config-client/app/router"
)

func Detect() router.Configurator {
	itb := intelbras.NewConfigurator()
	if itb.Detected() {
		fmt.Println("Intelbras router detected.")
		return itb
	}

	fmt.Println("Router not found.")
	return nil
}

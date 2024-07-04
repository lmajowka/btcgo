/**
 * App BTCGO
 * Versão Beta
 */

package main

import (
	app "btcgo/cmd/core"
	"btcgo/cmd/utils"
)

func main() {
	version := "v0.6b"

	utils.ClearConsole()
	utils.Title(version)
	app.NewApp()
}

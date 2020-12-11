/**
*  @file
*  @copyright defined in go-stem/LICENSE
 */

package main

import (
	"log"
	"os"

	"github.com/scdoproject/go-stem/cmd/client/cmd"
)

func main() {
	app := cmd.NewApp(true)

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

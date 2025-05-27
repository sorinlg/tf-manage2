package main

import (
	"os"

	"github.com/sorinlg/tf-manage2/internal/cli"
	"github.com/sorinlg/tf-manage2/internal/framework"
)

func main() {
	if err := cli.Execute(); err != nil {
		framework.Error(err.Error())
		os.Exit(1)
	}
}

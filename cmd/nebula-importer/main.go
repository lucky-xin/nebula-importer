package main

import (
	"github.com/lucky-xin/nebula-importer/pkg/cmd"
	"github.com/lucky-xin/nebula-importer/pkg/cmd/util"
)

func main() {
	command := cmd.NewDefaultImporterCommand()
	if err := util.Run(command); err != nil {
		util.CheckErr(err)
	}
}

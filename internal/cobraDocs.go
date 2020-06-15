package main

import (
	"github.com/spf13/cobra/doc"
	"github.com/syncromatics/kvetch/internal/cmd/kvetchctl"
)

func main() {
	err := doc.GenMarkdownTree(kvetchctl.RootCmd, "docs/kvetchctl")
	if err != nil {
		panic(err)
	}
}

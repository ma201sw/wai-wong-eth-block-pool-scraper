package main

import (
	"wai-wong/cmd/cmdroot"
)

func main() {
	if err := cmdroot.RootCmd.Execute(); err != nil {
		panic(err)
	}
}

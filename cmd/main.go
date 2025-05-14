package main

import f "github.com/aktagon/gofiles"

func main() {
	ui := f.NewFileExplorerUI()
	if err := ui.Start(); err != nil {
		panic(err)
	}
}

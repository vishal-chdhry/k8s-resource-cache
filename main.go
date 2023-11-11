package main

import (
	"fmt"
	"os"

	"github.com/vishal-chdhry/k8s-resource-cache/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

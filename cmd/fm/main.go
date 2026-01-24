package main

import (
	"os"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmd/root"
)

func main() {
	os.Exit(root.Execute())
}

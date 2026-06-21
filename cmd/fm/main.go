package main

import (
	"os"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmd/root"
	"github.com/marckohlbrugge/fastmail-cli/internal/jmap"
)

func main() {
	jmap.Version = root.Version
	os.Exit(root.Execute())
}

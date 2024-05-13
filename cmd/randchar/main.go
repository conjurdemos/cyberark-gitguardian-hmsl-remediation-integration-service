package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/conjurdemos/cyberark-gitguardian-hmsl-remediation-integration-service/brimstone/pkg/utils"
)

var (
	version string
)

func main() {
	num := flag.Int("n", 10, "number of chars to generate")
	list := flag.String("a", "abcdefghijklmnopqrstuvwxyz", "list of characters to use")
	ver := flag.Bool("version", false, "Print version")
	flag.Parse()

	if *ver {
		log.Printf("Version: %s\n", version)
		os.Exit(0)
	}

	seq := []rune(*list)
	fmt.Println(utils.RandSeq(seq, *num))
}

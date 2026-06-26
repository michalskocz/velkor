package main

import (
	"fmt"
	"log"
	"os"
	"velkor/configuration"
	"velkor/internal"

	"github.com/akamensky/argparse"
)

// Int and not bool because there can be more levels of debugging in feature
var debug int
var file os.File
var cfg configuration.Config

func parseArgs() {
	parser := argparse.NewParser("Local Action", "Local ci/cd tool")

	debug_leve := parser.Selector("d", "debug-level", []string{internal.DEBUG_ON_STR, internal.DEBUG_OFF_STR}, &argparse.Options{
		Required: false,
		Default:  internal.DEBUG_OFF_STR,
		Help:     "How meny info you got",
	})

	input_file := parser.File("f", "file", os.O_RDONLY, 0600, &argparse.Options{
		Required: false,
		Default:  internal.DEFAULT_INPUT_FILE,
		Help:     "Input config file",
	})

	if err := parser.Parse(os.Args); err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(-1)
	}

	if debug_leve != nil {
		switch *debug_leve {
		case internal.DEBUG_OFF_STR:
			debug = internal.DEBUG_OFF
		case internal.DEBUG_ON_STR:
			debug = internal.DEBUG_ON
		}
	}
	if input_file != nil {
		file = *input_file
	}

}

func getConfig() {
	data, err := os.ReadFile(file.Name())
	if err != nil {
		log.Fatal(err)
	}
	if conf, err := configuration.ParseYAMLConfig(data); err != nil {
		log.Fatal(err)
	} else if conf != nil {
		cfg = *conf
	} else {
		log.Fatal("Null pointer error")
	}

	if debug == internal.DEBUG_ON {
		fmt.Println(cfg.String())
	}
}

func main() {
	parseArgs()
	getConfig()
	if err := RunPipeline(); err != nil {
		log.Fatalf("Error in CI/CD: %v", err)
	}
}

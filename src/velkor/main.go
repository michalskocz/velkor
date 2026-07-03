/*
 Copyright (c) 2026 Michał Skoczylas

 Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

 1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

 2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

 THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

*/

package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/michalskocz/velkor/src/internal"

	"github.com/michalskocz/velkor/src/runner"

	"github.com/michalskocz/velkor/src/configuration"

	"github.com/akamensky/argparse"
)

// Int and not bool because there can be more levels of debugging in feature
var debug int
var file os.File
var cfg configuration.Config
var cpu int = 1
var goodIsolation bool = false

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

	cpu_per_container := parser.Int("c", "cpu", &argparse.Options{
		Required: false,
		Default:  cpu,
		Help:     "CPU cores per container",
		Validate: func(args []string) error {
			v, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("CPU must be an int")
			}
			if v <= 0 {
				return fmt.Errorf("CPU must be > 0")
			}
			return nil
		},
	})

	fast_flag := parser.Selector("i", "isolation", []string{"ON", "OFF"}, &argparse.Options{
		Required: false,
		Default:  "ON",
		Help:     "Uses flags that reduce security in exchange for performance. Default: ON",
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

	if fast_flag != nil {
		switch *fast_flag {
		case "ON":
			goodIsolation = true
		case "OF":
			goodIsolation = false
		}
	}

	if cpu_per_container != nil {
		cpu = *cpu_per_container
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
	if err := runner.RunPipeline(cfg, goodIsolation, cpu, debug); err != nil {
		log.Fatalf("Error in CI/CD: %v", err)
	}
}

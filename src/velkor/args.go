/*
 Copyright (c) 2026 Michał Skoczylas

 Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

 1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

 2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer
    in the documentation and/or other materials provided with the distribution.

 THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES,
 INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
 IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY,
 OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA,
 OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
 OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

*/

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/akamensky/argparse"
	"github.com/michalskocz/velkor/src/internal"
)

func parseArgs() {
	parser := argparse.NewParser("Local Action", "Local ci/cd tool")

	debug_leve := debugFlag(parser)
	input_file := configFlag(parser)
	cpu_per_container := cpuFlag(parser)
	isolation_flag := isolationFlag(parser)

	if err := parser.Parse(os.Args); err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(-1)
	}

	if err := parseDebug(debug_leve); err != nil {
		log.Fatal(err)
	}

	if err := parseConfig(input_file); err != nil {
		log.Fatal(err)
	}

	if err := parseIsolation(isolation_flag); err != nil {
		log.Fatal(err)
	}

	if err := parseCpu(cpu_per_container); err != nil {
		log.Fatal(err)
	}
}

func cpuFlag(parser *argparse.Parser) *int {
	const DEFAULT_CPU = 1

	return parser.Int("c", "cpu", &argparse.Options{
		Required: false,
		Default:  DEFAULT_CPU,
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
}

func debugFlag(parser *argparse.Parser) *string {
	return parser.Selector("d", "debug-level", []string{internal.DEBUG_ON_STR, internal.DEBUG_OFF_STR}, &argparse.Options{
		Required: false,
		Default:  internal.DEBUG_OFF_STR,
		Help:     "How meny info you got",
	})
}

func configFlag(parser *argparse.Parser) *os.File {
	return parser.File("f", "file", os.O_RDONLY, 0600, &argparse.Options{
		Required: false,
		Default:  internal.DEFAULT_INPUT_FILE,
		Help:     "Input config file",
	})
}

func isolationFlag(parser *argparse.Parser) *string {
	return parser.Selector("i", "isolation", []string{"ON", "OFF"}, &argparse.Options{
		Required: false,
		Default:  "ON",
		Help:     "Uses flags that reduce security in exchange for performance. Default: ON",
	})
}

var errNilPointer = errors.New("Argparser return nil pointer - unexpected behavior")
var errNewUnhadelCase = errors.New("Argparser configuration have new value with is unhandle")

func parseDebug(out *string) error {
	if out != nil {
		switch *out {
		case internal.DEBUG_OFF_STR:
			local.Debug = internal.DEBUG_OFF
		case internal.DEBUG_ON_STR:
			local.Debug = internal.DEBUG_ON
		default:
			return errNewUnhadelCase
		}
		return nil
	}
	return errNilPointer
}

func parseConfig(out *os.File) error {
	if out != nil {
		local.File = *out
		return nil
	}
	return errNilPointer
}

func parseIsolation(out *string) error {
	if out != nil {
		switch *out {
		case "ON":
			local.GoodIsolation = true
		case "OF":
			local.GoodIsolation = false
		default:
			return errNewUnhadelCase
		}
		return nil
	}
	return errNilPointer
}

func parseCpu(out *int) error {
	if out != nil {
		local.Cpu = *out
		return nil
	}
	return errNilPointer
}

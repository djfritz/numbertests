// Copyright 2025 David Fritz. All rights reserved.
// This software may be modified and distributed under the terms of the BSD
// 2-clause license. See the LICENSE file for details.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/djfritz/number"
)

const (
	internalPrecision = 34
)

var (
	fFile = flag.String("f", "", "Test file to run")
	fV    = flag.Bool("v", false, "verbose mode")
)

var (
	precision uint
	mode      int
	skip      bool
	testCount int
	success   int
	fail      int
	skipped   int
)

func main() {
	flag.Parse()

	f, err := os.Open(*fFile)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		process(strings.ToLower(scanner.Text()))
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	log.Printf("%v tests. %v successful, %v failed, %v skipped", testCount, success, fail, skipped)
}

func process(s string) {
	if s == "" {
		return
	} else if strings.HasPrefix(s, "--") {
		// comment
		return
	} else if strings.HasPrefix(s, "version") {
		return
	} else if strings.HasPrefix(s, "extended") {
		// doesn't apply to us
		return
	} else if strings.HasPrefix(s, "maxexponent") {
		// doesn't apply to us
		return
	} else if strings.HasPrefix(s, "minexponent") {
		// doesn't apply to us
		return
	} else if strings.HasPrefix(s, "precision") {
		processPrecision(s)
	} else if strings.HasPrefix(s, "rounding") {
		processRounding(s)
	} else {
		processTest(s)
	}
}

func processPrecision(s string) {
	s = strings.TrimSpace(strings.TrimPrefix(s, "precision:"))
	fields := strings.Fields(s)
	x, err := strconv.ParseUint(fields[0], 10, 64)
	if err != nil {
		log.Fatalf("parsing precision: %v, %v", err, s)
	}
	precision = uint(x)

	if *fV {
		fmt.Println("setting precision:", s)
	}
}

func processRounding(s string) {
	s = strings.TrimSpace(strings.TrimPrefix(s, "rounding:"))
	skip = false
	switch s {
	case "half_even":
		mode = number.ModeNearestEven
	case "half_up":
		mode = number.ModeNearest
	case "down", "zero":
		mode = number.ModeDown
	case "half_down", "floor", "ceiling", "up":
		skip = true
	default:
		log.Fatalf("invalid rounding mode: %v", s)
	}

	if *fV {
		fmt.Println("setting rounding mode:", s)
	}
}

func processTest(s string) {
	testCount++
	if skip {
		skipped++
		if *fV {
			log.Printf("skipping test: %v. Precision: %v. Rounding mode: %v", s, precision, mode)
		}
		return
	}

	fields := strings.Fields(s)

	if len(fields) < 6 {
		log.Fatalf("invalid input: %v", s)
	}
	name := fields[0]
	op := fields[1]
	l := strings.Trim(fields[2], "'")
	r := strings.Trim(fields[3], "'")
	e := strings.Trim(fields[5], "'")

	if *fV {
		fmt.Printf("test %v, op %v, l %v, r %v, expected %v", name, op, l, r, e)
	}

	lo, err := number.ParseReal(l, internalPrecision)
	if err != nil {
		log.Fatalf("parsing: %v: %v", l, err)
	}
	ro, err := number.ParseReal(r, internalPrecision)
	if err != nil {
		log.Fatalf("parsing: %v: %v", r, err)
	}
	ez, err := number.ParseReal(e, internalPrecision)
	if err != nil {
		log.Fatalf("parsing: %v: %v", e, err)
	}

	lo.SetMode(mode)
	ro.SetMode(mode)

	var z *number.Real
	switch op {
	case "add":
		z = lo.Add(ro)
	case "divide":
		z = lo.Div(ro)
	case "multiply":
		z = lo.Mul(ro)
	case "power":
		z = lo.Pow(ro)
	default:
		log.Fatalf("invalid op %v", op)
	}

	z.SetPrecision(precision)

	if *fV {
		log.Printf("result after rounding: %v", z)
	}

	if z.String() != ez.String() {
		fail++
		log.Printf("failed test: %v, %v != %v, precision: %v, rounding mode: %v", s, z, ez, precision, mode)
	} else {
		success++
	}
}

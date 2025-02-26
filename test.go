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

var (
	fV = flag.Bool("v", false, "verbose mode")
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

	files := flag.Args()

	for _, v := range files {
		f, err := os.Open(v)
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
		f.Close()
	}

	log.Printf("%v tests. %v successful, %v failed, %v skipped", testCount, success, fail, skipped)
}

func process(s string) {
	s = strings.TrimSpace(s)
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
	case "zero":
		mode = number.ModeZero
	case "half_down", "floor", "ceiling", "up", "down":
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

	var lo, ro, ez *number.Real
	var e string
	var err error
	if len(fields) < 5 {
		log.Fatalf("invalid input: %v", s)
	}

	name := fields[0]
	op := fields[1]
	l := strings.Trim(fields[2], "'")

	if l == "#" {
		skipped++
		if *fV {
			log.Printf("skipping test: %v. Precision: %v. Rounding mode: %v", s, precision, mode)
		}
		return
	}

	lo, err = number.ParseReal(l, uint(len(l))*2)
	if err != nil {
		log.Fatalf("parsing: %v: %v", l, err)
	}
	lo.SetMode(mode)
	lo.SetPrecision(precision)

	if len(fields) == 5 || fields[3] == "->" {
		// single operand
		e = strings.Trim(fields[4], "'")
		if e == "?" {
			skipped++
			if *fV {
				log.Printf("skipping test: %v. Precision: %v. Rounding mode: %v", s, precision, mode)
			}
			return
		}
		if *fV {
			fmt.Printf("test %v, op %v, l %v, expected %v", name, op, l, e)
		}
	} else {
		// dual operand
		e = strings.Trim(fields[5], "'")
		if e == "?" {
			skipped++
			if *fV {
				log.Printf("skipping test: %v. Precision: %v. Rounding mode: %v", s, precision, mode)
			}
			return
		}
		r := strings.Trim(fields[3], "'")
		ro, err = number.ParseReal(r, uint(len(r))*2)
		if err != nil {
			log.Fatalf("parsing: %v: %v", r, err)
		}
		ro.SetMode(mode)
		ro.SetPrecision(precision)
		if *fV {
			fmt.Printf("test %v, op %v, l %v, r %v, expected %v", name, op, l, r, e)
		}
	}

	ez, err = number.ParseReal(e, uint(len(e))*2)
	if err != nil {
		log.Fatalf("parsing: %v: %v", e, err)
	}

	var z *number.Real
	switch op {
	case "abs":
		z = lo.Abs()
	case "add":
		z = lo.Add(ro)
	case "subtract":
		z = lo.Sub(ro)
	case "divide":
		z = lo.Div(ro)
	case "multiply":
		z = lo.Mul(ro)
	case "power":
		z = lo.Pow(ro)
	case "exp":
		z = lo.Exp()
	case "ln":
		z = lo.Ln()
	case "squareroot":
		z = lo.Sqrt()
	case "compare":
		z = number.NewInt64(int64(lo.Compare(ro)))
	case "max":
		z = lo.Max(ro)
	case "min":
		z = lo.Min(ro)
	case "remainder":
		z = lo.Remainder(ro)
	default:
		if *fV {
			log.Printf("skipping test: %v. Precision: %v. Rounding mode: %v", s, precision, mode)
		}
		return
	}

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

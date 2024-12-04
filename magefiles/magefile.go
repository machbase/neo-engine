package main

import (
	"fmt"
	"os"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

func Test() error {
	mg.Deps(CheckTmp)
	if err := sh.RunV("go", "test", ".", "-count", "1"); err != nil {
		return err
	}
	fmt.Println("Test done.")
	return nil
}

func Bench() error {
	mg.Deps(CheckTmp)
	if err := sh.RunV("go", "test", "-benchmem", "-run", "^$", "-bench", "^Benchmark.*$", "github.com/machbase/neo-engine/v8", "-timeout", "60s", "-v"); err != nil {
		return err
	}
	fmt.Println("Benchmark done.")
	return nil
}

func CheckTmp() error {
	_, err := os.Stat("tmp")
	if err != nil && err != os.ErrNotExist {
		err = os.Mkdir("tmp", 0755)
	} else if err != nil && err == os.ErrExist {
		return nil
	}
	return err
}

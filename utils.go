package main

import (
	"fmt"

	"github.com/pkg/errors"
)

func showErrMsg() {
	if err := recover(); err != nil {
		fmt.Printf("%+v\n", err)
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
		// log.Fatal(err)
	}
}

func checkErrWithMsg(err error, msg string) {
	if err != nil {
		panic(errors.Wrap(err, msg))
	}
}

func Min(nums ...int64) int64 {
	var minNum int64 = 1<<15 - 1
	for _, num := range nums {
		if num < minNum {
			minNum = num
		}
	}
	return minNum
}

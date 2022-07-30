package main

import (
	"fmt"

	"github.com/go-ego/gse"
	"github.com/pkg/errors"
)

func TokenizeJp(text string) []string {
	var seg gse.Segmenter
	seg.SkipLog = true
	err := seg.LoadDict("jp")
	CheckErrWithMsg(err, "load dictionary error")
	segs := seg.Cut(text, true)
	return segs
}

func ShowErrMsg() {
	if err := recover(); err != nil {
		fmt.Printf("%+v\n", err)
	}
}

func CheckErr(err error) {
	if err != nil {
		panic(err)
		// log.Fatal(err)
	}
}

func CheckErrWithMsg(err error, msg string) {
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

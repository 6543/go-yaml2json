// SPDX-FileCopyrightText: 2023 6543 <6543@obermui.de>
// SPDX-License-Identifier: MIT

//go:build gofuzz
// +build gofuzz

/*
	How to fuzz:
	- go install  github.com/dvyukov/go-fuzz/go-fuzz@latest github.com/dvyukov/go-fuzz/go-fuzz-build@latest
	- go get github.com/dvyukov/go-fuzz/go-fuzz-dep
	- go-fuzz-build
	- go-fuzz -bin=./yaml2json-fuzz.zip
*/

package yaml2json

func Fuzz(data []byte) int {
	_, err := Convert(data)
	if err != nil {
		return 0
	}
	return 1
}

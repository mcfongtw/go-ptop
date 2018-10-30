package main

import (
	"fmt"
	"regexp"
)

func parseRegexByGroup(regEx, expr string) (paramsMap map[string]string) {

	var compRegEx = regexp.MustCompile(regEx)
	match := compRegEx.FindStringSubmatch(expr)

	paramsMap = make(map[string]string)
	for i, name := range compRegEx.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}
	return paramsMap
}

func stringify64BitAddress(addr uint64)(string) {
	hexAddr := "0x" + fmt.Sprintf("%016x", addr)

	return hexAddr
}
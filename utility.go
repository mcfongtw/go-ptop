package main

import (
	"fmt"
	"regexp"
)

func ParseRegexByGroup(regEx, expr string) (paramsMap map[string]string) {

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

func Stringify64BitAddress(addr uint64)(string) {
	hexAddr := "0x" + fmt.Sprintf("%016x", addr)

	return hexAddr
}

func StringfyInteger(val int) (string) {
	str := fmt.Sprintf("%v", val)

	return str
}

func StringfyUinteger32(val uint32) (string) {
	str := fmt.Sprintf("%v", val)

	return str
}

func StringfyUinteger64(val uint64) (string) {
	str := fmt.Sprintf("%v", val)

	return str
}
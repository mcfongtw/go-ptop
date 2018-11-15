package main

import (
	"fmt"
	"regexp"
	"strings"
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

func PrintMemorySegments(listOfMemorySegments *[]TaskMemorySegment) {
	fmt.Printf("[%-18s : %-18s] %9s %9s %9s %9s %9s %9s %9s %-10s %-30s\n", "START ADDR", "STOP ADDR", "PSS", "RSS", "DIRTY", "RD BYTES", "WRT BYTES", "RD CNT", "WRT CNT", "TYPE", "DATA")
	for i := 0; i < len(*listOfMemorySegments); i++ {
		segment := (*listOfMemorySegments)[i]

		if(strings.HasPrefix(segment.Path, "/")){
			fmt.Printf("[%-18v : %-18v] %9v %9v %9v %9v %9v %9v %9v [%-10s] %-30v\n", Stringify64BitAddress(segment.stackStart), Stringify64BitAddress(segment.stackStop), segment.Pss, segment.Rss, segment.PrivateDirty, segment.ReadBytes, segment.WriteBytes, segment.ReadCount, segment.WriteCount, segment.frameType, segment.Path)
		} else {
			fmt.Printf("[%-18v : %-18v] %9v %9v %9v %9v %9v %9v %9v [%-10s] %-30v\n", Stringify64BitAddress(segment.stackStart), Stringify64BitAddress(segment.stackStop), segment.Pss, segment.Rss, segment.PrivateDirty, segment.ReadBytes, segment.WriteBytes, segment.ReadCount, segment.WriteCount, segment.frameType, segment.Path)
		}

	}
}
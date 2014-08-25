package main

import (
	"fmt"
)

type MarkupAsciiDoc struct {
}

func (this *MarkupAsciiDoc) link(anchorName, linkText string) string {
	if linkText == "" {
		return fmt.Sprintf("<<%s,%s>>", anchorName, anchorName)
	}
	return fmt.Sprintf("<<%s,%s>>", anchorName, linkText)
}
func (this *MarkupAsciiDoc) tableHeader(tableTitle string) string {
	return fmt.Sprintf(".%s\n[width=\"80%%\",options=\"header\"]\n|==========\n", tableTitle)
}
func (this *MarkupAsciiDoc) tableFooter() string {
	return "|==========\n\n"
}

func (this *MarkupAsciiDoc) tableRow(args ...string) string {
	var retval string
	for _, arg := range args {
		retval += fmt.Sprintf("|%s ", arg)
	}
	return retval + "\n"
}

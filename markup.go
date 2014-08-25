package main

type Markup interface {
	link(anchorName, linkText string) string
	tableHeader(tableTitle string) string
	tableFooter() string
	tableRow(args ...string) string
}

package main

type Markup interface {
	anchor(anchorName string) string
	sectionHeader(level int, text string) string
	bulletedItem(level int, text string) string
	numberedItem(level int, text string) string
	link(anchorName, linkText string) string
	tableHeader(tableTitle string) string
	tableFooter() string
	tableRow(args ...string) string
}

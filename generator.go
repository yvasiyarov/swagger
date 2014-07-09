package main

import (
	"flag"
	"github.com/yvasiyarov/swagger/parser"
	"log"
	"os"
	"path"
)

var apiPackage = flag.String("apiPackage", "", "Package which implement API controllers")
var mainApiFile = flag.String("mainApiFile", "", "File with general API annotation, relatively to $GOPATH")

func main() {
	flag.Parse()
	if *apiPackage == "" || *mainApiFile == "" {
		flag.PrintDefaults()
		return
	}
	parser := parser.NewParser()

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		log.Fatalf("Please, set $GOPATH environment variable\n")
	}

	parser.ParseGeneralApiInfo(path.Join(gopath, "src", *mainApiFile))
	parser.ParseApiDescription(*apiPackage)

	log.Println(string(parser.GetResourceListingJson()))
	log.Println(string(parser.GetApiDescriptionJson()))
}

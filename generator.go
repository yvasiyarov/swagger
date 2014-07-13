package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"github.com/yvasiyarov/swagger/parser"
	"log"
	"os"
	"path"
	"strings"
)

var apiPackage = flag.String("apiPackage", "", "Package which implement API controllers")
var mainApiFile = flag.String("mainApiFile", "", "File with general API annotation, relatively to $GOPATH")
var basePath = flag.String("basePath", "http://127.0.0.1:3000", "Web service base path")

var generatedFileTemplate = `
package main

var resourceListingJson = {{resourceListing}}
var apiDescriptionsJson = {{apiDescriptions}}
`

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

	parser.BasePath = *basePath
	parser.ParseGeneralApiInfo(path.Join(gopath, "src", *mainApiFile))
	parser.ParseTypeDefinitions(*apiPackage)
	parser.ParseApiDescription(*apiPackage)

	//os.Mkdir(path.Join(curpath, "docs"), 0755)
	fd, err := os.Create(path.Join("./", "docs.go"))
	if err != nil {
		log.Fatalf("Can not create document file: %v\n", err)
	}
	defer fd.Close()

	var apiDescriptions bytes.Buffer
	for apiKey, apiDescription := range parser.TopLevelApis {
		apiDescriptions.WriteString("\"" + apiKey + "\":")

		apiDescriptions.WriteString("`")
		json, err := json.MarshalIndent(apiDescription, "", "    ")
		if err != nil {
			log.Fatalf("Can not serialise []ApiDescription to JSON: %v\n", err)
		}
		apiDescriptions.Write(json)
		apiDescriptions.WriteString("`,")
	}

	doc := strings.Replace(generatedFileTemplate, "{{resourceListing}}", "`"+string(parser.GetResourceListingJson())+"`", -1)
	doc = strings.Replace(doc, "{{apiDescriptions}}", "map[string]string{"+apiDescriptions.String()+"}", -1)

	fd.WriteString(doc)
	log.Println("Doc file generated")
	//log.Println(string(parser.GetResourceListingJson()))
	//log.Println(string(parser.GetApiDescriptionJson()))
}

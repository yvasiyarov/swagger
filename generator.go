package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/yvasiyarov/swagger/parser"
	"go/ast"
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
//This file is generated automatically. Dont try to edit it manually.

var resourceListingJson = {{resourceListing}}
var apiDescriptionsJson = {{apiDescriptions}}
`

func IsController(funcDeclaration *ast.FuncDecl) bool {
	if funcDeclaration.Recv != nil && len(funcDeclaration.Recv.List) > 0 {
		if starExpression, ok := funcDeclaration.Recv.List[0].Type.(*ast.StarExpr); ok {
			receiverName := fmt.Sprint(starExpression.X)
			return strings.Index(receiverName, "Context") != -1
		}
	}
	return false
}

func generateSwaggerDocs(parser *parser.Parser) {
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
}

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
	parser.IsController = IsController

	log.Println("Start parsing")
	parser.ParseGeneralApiInfo(path.Join(gopath, "src", *mainApiFile))
	parser.ParseTypeDefinitions(*apiPackage)
	parser.ParseApiDescription(*apiPackage)

	log.Println("Finish parsing")
	generateSwaggerDocs(parser)

	log.Println("Doc file generated")
}

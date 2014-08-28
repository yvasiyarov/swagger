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

const (
	AVAILABLE_FORMATS = "go|swagger|asciidoc|markdown|confluence"
)

var apiPackage = flag.String("apiPackage", "", "The package that implements the API controllers, relative to $GOPATH/src")
var mainApiFile = flag.String("mainApiFile", "", "The file that contains the general API annotations, relative to $GOPATH/src")
var basePath = flag.String("basePath", "http://127.0.0.1:3000", "Web service base path")
var outputFormat = flag.String("format", "go", "Output format type for the generated files: "+AVAILABLE_FORMATS)
var outputSpec = flag.String("output", "", "Output (path) for the generated file(s)")

var generatedFileTemplate = `
package main
//This file is generated automatically. Do not try to edit it manually.

var resourceListingJson = {{resourceListing}}
var apiDescriptionsJson = {{apiDescriptions}}
`

// It must return true if funcDeclaration is controller. We will try to parse only comments before controllers
func IsController(funcDeclaration *ast.FuncDecl) bool {
	if funcDeclaration.Recv != nil && len(funcDeclaration.Recv.List) > 0 {
		if starExpression, ok := funcDeclaration.Recv.List[0].Type.(*ast.StarExpr); ok {
			receiverName := fmt.Sprint(starExpression.X)
			return strings.Index(receiverName, "Context") != -1 || strings.Index(receiverName, "Controller") != -1
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

func shortModelName(longModelName string) string {
	parts := strings.Split(longModelName, ".")
	return parts[len(parts)-1]
}

func generateSwaggerUiFiles(parser *parser.Parser) {
	fd, err := os.Create(path.Join(*outputSpec, "index.json"))
	if err != nil {
		log.Fatalf("Can not create the master index.json file: %v\n", err)
	}
	defer fd.Close()
	fd.WriteString(string(parser.GetResourceListingJson()))

	for apiKey, apiDescription := range parser.TopLevelApis {
		err = os.MkdirAll(path.Join(*outputSpec, apiKey), 0777)
		fd, err = os.Create(path.Join(*outputSpec, apiKey, "index.json"))
		if err != nil {
			log.Fatalf("Can not create the %s/index.json file: %v\n", apiKey, err)
		}
		defer fd.Close()
		json, err := json.MarshalIndent(apiDescription, "", "    ")
		if err != nil {
			log.Fatalf("Can not serialise []ApiDescription to JSON: %v\n", err)
		}
		fd.Write(json)
		log.Printf("Wrote %v/index.json", apiKey)
	}
}

func InitParser() *parser.Parser {
	parser := parser.NewParser()

	parser.BasePath = *basePath
	parser.IsController = IsController

	parser.TypesImplementingMarshalInterface["NullString"] = "string"
	parser.TypesImplementingMarshalInterface["NullInt64"] = "int"
	parser.TypesImplementingMarshalInterface["NullFloat64"] = "float"
	parser.TypesImplementingMarshalInterface["NullBool"] = "bool"

	return parser
}

func main() {
	flag.Parse()

	if *mainApiFile == "" {
		*mainApiFile = *apiPackage + "/main.go"
	}
	if *apiPackage == "" {
		flag.PrintDefaults()
		return
	}

	parser := InitParser()
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		log.Fatalf("Please, set $GOPATH environment variable\n")
	}

	log.Println("Start parsing")
	parser.ParseGeneralApiInfo(path.Join(gopath, "src", *mainApiFile))
	parser.ParseApi(*apiPackage)
	log.Println("Finish parsing")

	format := strings.ToLower(*outputFormat)
	switch format {
	case "go":
		generateSwaggerDocs(parser)
		log.Println("Doc file generated")
	case "asciidoc":
		markupAsciiDoc := new(MarkupAsciiDoc)
		generateMarkup(parser, markupAsciiDoc, ".adoc")
		log.Println("AsciiDoc file generated")
	case "markdown":
		markupMarkDown := new(MarkupMarkDown)
		generateMarkup(parser, markupMarkDown, ".md")
		log.Println("MarkDown file generated")
	case "confluence":
		markupConfluence := new(MarkupConfluence)
		generateMarkup(parser, markupConfluence, ".confluence")
		log.Println("Confluence file generated")
	case "swagger":
		generateSwaggerUiFiles(parser)
		log.Println("Swagger UI files generated")
	default:
		log.Fatalf("Invalid -format specified. Must be one of %v.", AVAILABLE_FORMATS)
	}

}

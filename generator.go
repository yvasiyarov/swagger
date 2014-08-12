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
var swaggerUiPath = flag.String("swaggerUiPath", "", "(Optional) output path for Swagger UI files (instead of docs.go)")

var generatedFileTemplate = `
package main
//This file is generated automatically. Dont try to edit it manually.

var resourceListingJson = {{resourceListing}}
var apiDescriptionsJson = {{apiDescriptions}}
`

// It must return true if funcDeclaration is controller. We will try to parse only comments before controllers
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

func generateSwaggerUiFiles(parser *parser.Parser) {
	fd, err := os.Create(path.Join(*swaggerUiPath, "index.json"))
	if err != nil {
		log.Fatalf("Can not create the master index.json file: %v\n", err)
	}
	defer fd.Close()
	fd.WriteString(string(parser.GetResourceListingJson()))

	for apiKey, apiDescription := range parser.TopLevelApis {
		err = os.MkdirAll(path.Join(*swaggerUiPath, apiKey), 0777)
		fd, err = os.Create(path.Join(*swaggerUiPath, apiKey, "index.json"))
		if err != nil {
			log.Fatalf("Can not create the %s/index.json file: %v\n", apiKey, err)
		}
		defer fd.Close()
		json, err := json.MarshalIndent(apiDescription, "", "    ")
		if err != nil {
			log.Fatalf("Can not serialise []ApiDescription to JSON: %v\n", err)
		}
		fd.Write(json)
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
	if *swaggerUiPath == "" {
		generateSwaggerDocs(parser)
	} else {
		generateSwaggerUiFiles(parser)
	}

	log.Println("Doc file generated")
}

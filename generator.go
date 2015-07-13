package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/yvasiyarov/swagger/markup"
	"github.com/yvasiyarov/swagger/parser"
)

const (
	AVAILABLE_FORMATS = "go|swagger|asciidoc|markdown|confluence"
)

var apiPackage = flag.String("apiPackage", "", "The package that implements the API controllers, relative to $GOPATH/src")
var mainApiFile = flag.String("mainApiFile", "", "The file that contains the general API annotations, relative to $GOPATH/src")
var outputFormat = flag.String("format", "go", "Output format type for the generated files: "+AVAILABLE_FORMATS)
var outputSpec = flag.String("output", "", "Output (path) for the generated file(s)")
var controllerClass = flag.String("controllerClass", "", "Speed up parsing by specifying which receiver objects have the controller methods")

var generatedFileTemplate = `
package main
//This file is generated automatically. Do not try to edit it manually.

var resourceListingJson = {{resourceListing}}
var apiDescriptionsJson = {{apiDescriptions}}
`

// It must return true if funcDeclaration is controller. We will try to parse only comments before controllers
func IsController(funcDeclaration *ast.FuncDecl) bool {
	if len(*controllerClass) == 0 {
		// Search every method
		return true
	}
	if funcDeclaration.Recv != nil && len(funcDeclaration.Recv.List) > 0 {
		if starExpression, ok := funcDeclaration.Recv.List[0].Type.(*ast.StarExpr); ok {
			receiverName := fmt.Sprint(starExpression.X)
			matched, err := regexp.MatchString(string(*controllerClass), receiverName)
			if err != nil {
				log.Fatalf("The -controllerClass argument is not a valid regular expression: %v\n", err)
			}
			return matched
		}
	}
	return false
}

func generateSwaggerDocs(parser *parser.Parser) error {
	fd, err := os.Create(path.Join(*outputSpec, "docs.go"))
	if err != nil {
		return fmt.Errorf("Can not create document file: %v\n", err)
	}
	defer fd.Close()

	var apiDescriptions bytes.Buffer
	for apiKey, apiDescription := range parser.TopLevelApis {
		apiDescriptions.WriteString("\"" + apiKey + "\":")

		apiDescriptions.WriteString("`")
		json, err := json.MarshalIndent(apiDescription, "", "    ")
		if err != nil {
			return fmt.Errorf("Can not serialise []ApiDescription to JSON: %v\n", err)
		}
		apiDescriptions.Write(json)
		apiDescriptions.WriteString("`,")
	}

	doc := strings.Replace(generatedFileTemplate, "{{resourceListing}}", "`"+string(parser.GetResourceListingJson())+"`", -1)
	doc = strings.Replace(doc, "{{apiDescriptions}}", "map[string]string{"+apiDescriptions.String()+"}", -1)

	fd.WriteString(doc)

	return nil
}

func generateSwaggerUiFiles(parser *parser.Parser) error {
	fd, err := os.Create(path.Join(*outputSpec, "index.json"))
	if err != nil {
		return fmt.Errorf("Can not create the master index.json file: %v\n", err)
	}
	defer fd.Close()
	fd.WriteString(string(parser.GetResourceListingJson()))

	for apiKey, apiDescription := range parser.TopLevelApis {
		err = os.MkdirAll(path.Join(*outputSpec, apiKey), 0777)
		if err != nil {
			return err
		}

		fd, err = os.Create(path.Join(*outputSpec, apiKey, "index.json"))
		if err != nil {
			return fmt.Errorf("Can not create the %s/index.json file: %v\n", apiKey, err)
		}
		defer fd.Close()

		json, err := json.MarshalIndent(apiDescription, "", "    ")
		if err != nil {
			return fmt.Errorf("Can not serialise []ApiDescription to JSON: %v\n", err)
		}

		fd.Write(json)
		log.Printf("Wrote %v/index.json", apiKey)
	}

	return nil
}

func InitParser() *parser.Parser {
	parser := parser.NewParser()

	parser.BasePath = "{{.}}"
	parser.IsController = IsController

	parser.TypesImplementingMarshalInterface["NullString"] = "string"
	parser.TypesImplementingMarshalInterface["NullInt64"] = "int"
	parser.TypesImplementingMarshalInterface["NullFloat64"] = "float"
	parser.TypesImplementingMarshalInterface["NullBool"] = "bool"

	return parser
}

type GeneratorParams struct {
	ApiPackage, MainApiFile, OutputFormat, OutputSpec, ControllerClass string
}

func Generate(params GeneratorParams) error {
	parser := InitParser()
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return errors.New("Please, set $GOPATH environment variable\n")
	}

	log.Println("Start parsing")

	//Support gopaths with multiple directories
	dirs := strings.Split(gopath, ":")
	found := false
	for _, d := range dirs {
		apifile := path.Join(d, "src", params.MainApiFile)
		if _, err := os.Stat(apifile); err == nil {
			parser.ParseGeneralApiInfo(apifile)
			found = true
		}
	}
	if found == false {
		apifile := path.Join(gopath, "src", params.MainApiFile)
		return fmt.Errorf("Could not find apifile %s to parse\n", apifile)
	}

	parser.ParseApi(params.ApiPackage)
	log.Println("Finish parsing")

	var err error
	confirmMsg := ""
	format := strings.ToLower(params.OutputFormat)
	switch format {
	case "go":
		err = generateSwaggerDocs(parser)
		confirmMsg = "Doc file generated"
	case "asciidoc":
		err = markup.GenerateMarkup(parser, new(markup.MarkupAsciiDoc), &params.OutputSpec, ".adoc")
		confirmMsg = "AsciiDoc file generated"
	case "markdown":
		err = markup.GenerateMarkup(parser, new(markup.MarkupMarkDown), &params.OutputSpec, ".md")
		confirmMsg = "MarkDown file generated"
	case "confluence":
		err = markup.GenerateMarkup(parser, new(markup.MarkupConfluence), &params.OutputSpec, ".confluence")
		confirmMsg = "Confluence file generated"
	case "swagger":
		err = generateSwaggerUiFiles(parser)
		confirmMsg = "Swagger UI files generated"
	default:
		err = fmt.Errorf("Invalid -format specified. Must be one of %v.", AVAILABLE_FORMATS)
	}

	if err != nil {
		return err
	}
	log.Println(confirmMsg)

	return nil
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

	params := GeneratorParams{
		ApiPackage:      *apiPackage,
		MainApiFile:     *mainApiFile,
		OutputFormat:    *outputFormat,
		OutputSpec:      *outputSpec,
		ControllerClass: *controllerClass,
	}

	err := Generate(params)
	if err != nil {
		log.Fatal(err.Error())
	}
}

package generator

import (
	"bytes"
	"encoding/json"
<<<<<<< HEAD:generator.go
	"flag"
=======
	"errors"
>>>>>>> yvasiyarov/swagger2.0:generator/generator.go
	"fmt"
	"github.com/yvasiyarov/swagger/markup"
	"github.com/yvasiyarov/swagger/parser"
	"go/ast"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
)

const (
	AVAILABLE_FORMATS = "go|swagger|asciidoc|markdown|confluence"
)

<<<<<<< HEAD:generator.go
var apiPackage = flag.String("apiPackage", "", "The package that implements the API controllers, relative to $GOPATH/src")
var mainApiFile = flag.String("mainApiFile", "", "The file that contains the general API annotations, relative to $GOPATH/src")
var basePath = flag.String("basePath", "http://127.0.0.1:3000", "Web service base path")
var outputFormat = flag.String("format", "go", "Output format type for the generated files: "+AVAILABLE_FORMATS)
var outputSpec = flag.String("output", "", "Output (path) for the generated file(s)")
var controllerClass = flag.String("controllerClass", "", "Speed up parsing by specifying which receiver objects have the controller methods")

=======
>>>>>>> yvasiyarov/swagger2.0:generator/generator.go
var generatedFileTemplate = `
package main
//This file is generated automatically. Do not try to edit it manually.

var resourceListingJson = {{resourceListing}}
var apiDescriptionsJson = {{apiDescriptions}}
`

// It must return true if funcDeclaration is controller. We will try to parse only comments before controllers
func IsController(funcDeclaration *ast.FuncDecl, controllerClass string) bool {
	if len(controllerClass) == 0 {
		// Search every method
		return true
	}
	if funcDeclaration.Recv != nil && len(funcDeclaration.Recv.List) > 0 {
		if starExpression, ok := funcDeclaration.Recv.List[0].Type.(*ast.StarExpr); ok {
			receiverName := fmt.Sprint(starExpression.X)
			matched, err := regexp.MatchString(string(controllerClass), receiverName)
			if err != nil {
				log.Fatalf("The -controllerClass argument is not a valid regular expression: %v\n", err)
			}
			return matched
		}
	}
	return false
}

<<<<<<< HEAD:generator.go
func generateSwaggerDocs(parser *parser.Parser) {
	fd, err := os.Create(path.Join(*outputSpec, "docs.go"))
=======
func generateSwaggerDocs(parser *parser.Parser, outputSpec string) error {
	fd, err := os.Create(path.Join(outputSpec, "docs.go"))
>>>>>>> yvasiyarov/swagger2.0:generator/generator.go
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

<<<<<<< HEAD:generator.go
func generateSwaggerUiFiles(parser *parser.Parser) {
	fd, err := os.Create(path.Join(*outputSpec, "index.json"))
=======
func generateSwaggerUiFiles(parser *parser.Parser, outputSpec string) error {
	fd, err := os.Create(path.Join(outputSpec, "index.json"))
>>>>>>> yvasiyarov/swagger2.0:generator/generator.go
	if err != nil {
		log.Fatalf("Can not create the master index.json file: %v\n", err)
	}
	defer fd.Close()
	fd.WriteString(string(parser.GetResourceListingJson()))

	for apiKey, apiDescription := range parser.TopLevelApis {
<<<<<<< HEAD:generator.go
		err = os.MkdirAll(path.Join(*outputSpec, apiKey), 0777)
		fd, err = os.Create(path.Join(*outputSpec, apiKey, "index.json"))
=======
		err = os.MkdirAll(path.Join(outputSpec, apiKey), 0777)
		if err != nil {
			return err
		}

		fd, err = os.Create(path.Join(outputSpec, apiKey, "index.json"))
>>>>>>> yvasiyarov/swagger2.0:generator/generator.go
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

func InitParser(controllerClass, ignore string) *parser.Parser {
	parser := parser.NewParser()

<<<<<<< HEAD:generator.go
	parser.BasePath = *basePath
=======
	parser.ControllerClass = controllerClass
>>>>>>> yvasiyarov/swagger2.0:generator/generator.go
	parser.IsController = IsController
	parser.Ignore = ignore

	parser.TypesImplementingMarshalInterface["NullString"] = "string"
	parser.TypesImplementingMarshalInterface["NullInt64"] = "int"
	parser.TypesImplementingMarshalInterface["NullFloat64"] = "float"
	parser.TypesImplementingMarshalInterface["NullBool"] = "bool"

	return parser
}

<<<<<<< HEAD:generator.go
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
=======
type Params struct {
	ApiPackage, MainApiFile, OutputFormat, OutputSpec, ControllerClass, Ignore string
}

func Run(params Params) error {
	parser := InitParser(params.ControllerClass, params.Ignore)
>>>>>>> yvasiyarov/swagger2.0:generator/generator.go
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		log.Fatalf("Please, set $GOPATH environment variable\n")
	}

	log.Println("Start parsing")

	//Support gopaths with multiple directories
	dirs := strings.Split(gopath, ":")
	found := false
	for _, d := range dirs {
		apifile := path.Join(d, "src", *mainApiFile)
		if _, err := os.Stat(apifile); err == nil {
			parser.ParseGeneralApiInfo(apifile)
			found = true
		}
	}
	if found == false {
		apifile := path.Join(gopath, "src", *mainApiFile)
		f, _ := fmt.Printf("Could not find apifile %s to parse\n", apifile)
		panic(f)
	}

	parser.ParseApi(*apiPackage)
	log.Println("Finish parsing")

	format := strings.ToLower(*outputFormat)
	switch format {
	case "go":
<<<<<<< HEAD:generator.go
		generateSwaggerDocs(parser)
		log.Println("Doc file generated")
=======
		err = generateSwaggerDocs(parser, params.OutputSpec)
		confirmMsg = "Doc file generated"
>>>>>>> yvasiyarov/swagger2.0:generator/generator.go
	case "asciidoc":
		markup.GenerateMarkup(parser, new(markup.MarkupAsciiDoc), outputSpec, ".adoc")
		log.Println("AsciiDoc file generated")
	case "markdown":
		markup.GenerateMarkup(parser, new(markup.MarkupMarkDown), outputSpec, ".md")
		log.Println("MarkDown file generated")
	case "confluence":
		markup.GenerateMarkup(parser, new(markup.MarkupConfluence), outputSpec, ".confluence")
		log.Println("Confluence file generated")
	case "swagger":
<<<<<<< HEAD:generator.go
		generateSwaggerUiFiles(parser)
		log.Println("Swagger UI files generated")
	default:
		log.Fatalf("Invalid -format specified. Must be one of %v.", AVAILABLE_FORMATS)
	}

=======
		err = generateSwaggerUiFiles(parser, params.OutputSpec)
		confirmMsg = "Swagger UI files generated"
	default:
		err = fmt.Errorf("Invalid -format specified. Must be one of %v.", AVAILABLE_FORMATS)
	}

	if err != nil {
		return err
	}
	log.Println(confirmMsg)

	return nil
>>>>>>> yvasiyarov/swagger2.0:generator/generator.go
}

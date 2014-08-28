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
	"sort"
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

func alphabeticalKeysOfSubApis(refs []*parser.ApiRef) ([]string, map[string]int) {
	index := map[string]int{}
	keys := make([]string, len(refs))
	for i, ref := range refs {
		subApiKey := ref.Path[1:]
		keys[i] = subApiKey
		index[subApiKey] = i
	}
	sort.Strings(keys)
	return keys, index
}
func alphabeticalKeysOfApiDeclaration(m map[string]*parser.ApiDeclaration) []string {
	keys := make([]string, len(m))
	i := 0
	for key, _ := range m {
		keys[i] = key
		i++
	}
	sort.Strings(keys)
	return keys
}

func generateMarkup(parser *parser.Parser, markup Markup, fileExtension string) {
	var filename string
	if *outputSpec == "" {
		filename = path.Join("./", "API", fileExtension)
	} else {
		filename = path.Join(*outputSpec)
	}
	fd, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Can not create document file: %v\n", err)
	}
	defer fd.Close()

	var buf bytes.Buffer

	/***************************************************************
	* Overall API
	***************************************************************/
	buf.WriteString(markup.sectionHeader(1, parser.Listing.Infos.Title))
	buf.WriteString(fmt.Sprintf("%s\n\n", parser.Listing.Infos.Description))

	/***************************************************************
	* Table of Contents (List of Sub-APIs)
	***************************************************************/
	buf.WriteString("Table of Contents\n\n")
	subApiKeys, subApiKeyIndex := alphabeticalKeysOfSubApis(parser.Listing.Apis)
	for _, subApiKey := range subApiKeys {
		buf.WriteString(markup.numberedItem(1, markup.link(subApiKey, parser.Listing.Apis[subApiKeyIndex[subApiKey]].Description)))
	}
	buf.WriteString("\n")

	for _, apiKey := range alphabeticalKeysOfApiDeclaration(parser.TopLevelApis) {

		apiDescription := parser.TopLevelApis[apiKey]
		/***************************************************************
		* Sub-API Specifications
		***************************************************************/
		buf.WriteString(markup.anchor(apiKey))
		buf.WriteString(markup.sectionHeader(2, apiKey))

		buf.WriteString(markup.tableHeader(""))
		buf.WriteString(markup.tableHeaderRow("Specification", "Value"))
		buf.WriteString(markup.tableRow("Resource Path", apiDescription.ResourcePath))
		buf.WriteString(markup.tableRow("API Version", apiDescription.ApiVersion))
		buf.WriteString(markup.tableRow("BasePath for the API", apiDescription.BasePath))
		buf.WriteString(markup.tableRow("Consumes", strings.Join(apiDescription.Consumes, ", ")))
		buf.WriteString(markup.tableRow("Produces", strings.Join(apiDescription.Produces, ", ")))
		buf.WriteString(markup.tableFooter())

		/***************************************************************
		* Sub-API Operations (Summary)
		***************************************************************/
		buf.WriteString("\n")
		buf.WriteString(markup.sectionHeader(3, "Operations"))
		buf.WriteString("\n")

		buf.WriteString(markup.tableHeader(""))
		buf.WriteString(markup.tableHeaderRow("Resource Path", "Operation", "Description"))
		for _, subapi := range apiDescription.Apis {
			for _, op := range subapi.Operations {
				pathString := strings.Replace(strings.Replace(subapi.Path, "{", "\\{", -1), "}", "\\}", -1)
				buf.WriteString(markup.tableRow(pathString, markup.link(op.Nickname, op.HttpMethod), op.Summary))
			}
		}
		buf.WriteString(markup.tableFooter())
		buf.WriteString("\n")

		/***************************************************************
		* Sub-API Operations (Details)
		***************************************************************/
		for _, subapi := range apiDescription.Apis {
			for _, op := range subapi.Operations {
				buf.WriteString("\n")
				operationString := fmt.Sprintf("%s (%s)", strings.Replace(strings.Replace(subapi.Path, "{", "\\{", -1), "}", "\\}", -1), op.HttpMethod)
				buf.WriteString(markup.anchor(op.Nickname))
				buf.WriteString(markup.sectionHeader(4, "API: "+operationString))
				buf.WriteString("\n\n" + op.Summary + "\n\n\n")

				if len(op.Parameters) > 0 {
					buf.WriteString(markup.tableHeader(""))
					buf.WriteString(markup.tableHeaderRow("Param Name", "Param Type", "Data Type", "Description", "Required?"))
					for _, param := range op.Parameters {
						isRequired := ""
						if param.Required {
							isRequired = "Yes"
						}
						buf.WriteString(markup.tableRow(param.Name, param.ParamType, param.DataType, param.Description, isRequired))
					}
					buf.WriteString(markup.tableFooter())
				}

				if len(op.ResponseMessages) > 0 {
					buf.WriteString(markup.tableHeader(""))
					buf.WriteString(markup.tableHeaderRow("Code", "Message", "Model"))
					for _, msg := range op.ResponseMessages {
						shortName := shortModelName(msg.ResponseModel)
						modelText := shortName
						if msg.ResponseModel != shortName {
							modelText = markup.link(msg.ResponseModel, shortName)
						}
						buf.WriteString(markup.tableRow(fmt.Sprintf("%v", msg.Code), msg.Message, modelText))
					}
					buf.WriteString(markup.tableFooter())
				}
			}
		}
		buf.WriteString("\n")

		/***************************************************************
		* Models
		***************************************************************/
		buf.WriteString("\n")
		buf.WriteString(markup.sectionHeader(3, "Models"))
		buf.WriteString("\n")

		for modelKey, model := range apiDescription.Models {
			buf.WriteString(markup.anchor(modelKey))
			buf.WriteString(markup.sectionHeader(4, shortModelName(modelKey)))
			buf.WriteString(markup.tableHeader(""))
			buf.WriteString(markup.tableHeaderRow("Field Name", "Field Type", "Description"))
			for fieldName, fieldProps := range model.Properties {
				buf.WriteString(markup.tableRow(fieldName, fieldProps.Type, fieldProps.Description))
			}
			buf.WriteString(markup.tableFooter())
			buf.WriteString("\nNote: These fields are listed in random order (for now).\n")
		}
		buf.WriteString("\n")

	}

	fd.WriteString(buf.String())
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

package generator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/yvasiyarov/swagger/markup"
	"github.com/yvasiyarov/swagger/parser"
)

const (
	AVAILABLE_FORMATS = "go|gopkg|swagger|asciidoc|markdown|confluence"
)

var (
	log = logrus.WithField("pkg", "generator")

	generatedFileTemplate = `
package main
//This file is generated automatically. Do not try to edit it manually.

var resourceListingJson = {{resourceListing}}
var apiDescriptionsJson = {{apiDescriptions}}
`

	generatedPkgTemplate = `
package {{packageName}}
//This file is generated automatically. Do not try to edit it manually.

var ResourceListingJson = {{resourceListing}}
var ApiDescriptionsJson = {{apiDescriptions}}
`
)

func generateSwaggerDocs(parser *parser.Parser, outputSpec string, pkg bool) error {
	fd, err := os.Create(path.Join(outputSpec, "docs.go"))
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

	var doc string
	if pkg {
		doc = strings.Replace(generatedPkgTemplate, "{{resourceListing}}", "`"+string(parser.GetResourceListingJson())+"`", -1)
		doc = strings.Replace(doc, "{{apiDescriptions}}", "map[string]string{"+apiDescriptions.String()+"}", -1)
		packageName := strings.Split(outputSpec, "/")
		doc = strings.Replace(doc, "{{packageName}}", packageName[len(packageName)-1], -1)
	} else {
		doc = strings.Replace(generatedFileTemplate, "{{resourceListing}}", "`"+string(parser.GetResourceListingJson())+"`", -1)
		doc = strings.Replace(doc, "{{apiDescriptions}}", "map[string]string{"+apiDescriptions.String()+"}", -1)
	}

	fd.WriteString(doc)

	return nil
}

func generateSwaggerUiFiles(parser *parser.Parser, outputSpec string) error {
	fd, err := os.Create(path.Join(outputSpec, "index.json"))
	if err != nil {
		return fmt.Errorf("Can not create the master index.json file: %v\n", err)
	}
	defer fd.Close()
	fd.WriteString(string(parser.GetResourceListingJson()))

	for apiKey, apiDescription := range parser.TopLevelApis {
		err = os.MkdirAll(path.Join(outputSpec, apiKey), 0777)
		if err != nil {
			return err
		}

		fd, err = os.Create(path.Join(outputSpec, apiKey, "index.json"))
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

type Params struct {
	ApiPackage, MainApiFile, OutputFormat, OutputSpec, ControllerClass, Ignore, VendoringPath string
	ContentsTable, Models, DisableVendoring                                                   bool
}

func Run(params Params) error {
	parser, err := parser.NewParser(params.ApiPackage, params.ControllerClass, params.Ignore,
		params.VendoringPath, params.DisableVendoring)
	if err != nil {
		return fmt.Errorf("Unable to initialize parser: %v", err)
	}

	log.Println("Start parsing")

	//Support gopaths with multiple directories
	dirs := strings.Split(parser.GoPath, ":")
	if runtime.GOOS == "windows" {
		dirs = strings.Split(parser.GoPath, ";")
	}

	found := false
	for _, d := range dirs {
		apifile := path.Join(d, "src", params.MainApiFile)
		if _, err := os.Stat(apifile); err == nil {
			parser.ParseGeneralApiInfo(apifile)

			log.Debugf("Found entry point API file '%v'", apifile)
			found = true
			break // file found, exit the loop
		}
	}
	if found == false {
		if _, err := os.Stat(params.MainApiFile); err == nil {
			parser.ParseGeneralApiInfo(params.MainApiFile)
		} else {
			apifile := path.Join(parser.GoPath, "src", params.MainApiFile)
			return fmt.Errorf("Could not find apifile %s to parse\n", apifile)
		}
	}

	parser.ParseApi()

	log.Println("Finish parsing")

	var confirmMsg string

	format := strings.ToLower(params.OutputFormat)

	switch format {
	case "go":
		err = generateSwaggerDocs(parser, params.OutputSpec, false)
		confirmMsg = "Doc file generated"
	case "gopkg":
		err = generateSwaggerDocs(parser, params.OutputSpec, true)
		confirmMsg = "Doc package generated"
	case "asciidoc":
		err = markup.GenerateMarkup(parser, new(markup.MarkupAsciiDoc), &params.OutputSpec, ".adoc", params.ContentsTable, params.Models)
		confirmMsg = "AsciiDoc file generated"
	case "markdown":
		err = markup.GenerateMarkup(parser, new(markup.MarkupMarkDown), &params.OutputSpec, ".md", params.ContentsTable, params.Models)
		confirmMsg = "MarkDown file generated"
	case "confluence":
		err = markup.GenerateMarkup(parser, new(markup.MarkupConfluence), &params.OutputSpec, ".confluence", params.ContentsTable, params.Models)
		confirmMsg = "Confluence file generated"
	case "swagger":
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
}

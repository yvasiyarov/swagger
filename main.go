package main

import (
	"flag"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/yvasiyarov/swagger/generator"
)

var apiPackage = flag.String("apiPackage", "", "The package that implements the API controllers, relative to $GOPATH/src")
var mainApiFile = flag.String("mainApiFile", "", "The file that contains the general API annotations, relative to $GOPATH/src")
var outputFormat = flag.String("format", "go", "Output format type for the generated files: "+generator.AVAILABLE_FORMATS)
var outputSpec = flag.String("output", "", "Output (path) for the generated file(s)")
var controllerClass = flag.String("controllerClass", "", "Speed up parsing by specifying which receiver objects have the controller methods")
var ignore = flag.String("ignore", "^$", "Ignore packages that satisfy this match")
var contentsTable = flag.Bool("contentsTable", true, "Generate the section Table of Contents")
var models = flag.Bool("models", true, "Generate the section models if any defined")
var vendoringPath = flag.String("vendoringPath", "", "Override default vendor directory")
var disableVendoring = flag.Bool("disableVendoring", false, "Disable vendor dir usage")
var enableDebug = flag.Bool("enableDebug", false, "Enable debug log output")

func init() {
	flag.Parse()

	if *enableDebug {
		log.SetLevel(log.DebugLevel)
		log.Info("Debug logging enabled")
	}
}

func main() {
	if *mainApiFile == "" {
		*mainApiFile = *apiPackage + "/main.go"
		log.Debugf("Using '%v' as main API file", *mainApiFile)
	}

	if *apiPackage == "" {
		flag.PrintDefaults()
		return
	}

	// Get rid of trailing /
	*vendoringPath = strings.TrimSuffix(*vendoringPath, "/")

	params := generator.Params{
		ApiPackage:       *apiPackage,
		MainApiFile:      *mainApiFile,
		OutputFormat:     *outputFormat,
		OutputSpec:       *outputSpec,
		ControllerClass:  *controllerClass,
		Ignore:           *ignore,
		ContentsTable:    *contentsTable,
		Models:           *models,
		VendoringPath:    *vendoringPath,
		DisableVendoring: *disableVendoring,
	}

	err := generator.Run(params)
	if err != nil {
		log.Fatal(err.Error())
	}
}

package parser

import (
	"encoding/json"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

type Parser struct {
	Listing          *ResourceListing
	TopLevelApis     map[string]*ApiDeclaration
	PackagesCache    map[string]map[string]*ast.Package
	CurrentPackage   string
	TypeDefinitions  map[string]map[string]*ast.TypeSpec
	PackagePathCache map[string]string
	PackageImports   map[string]map[string]string
	BasePath         string
	IsController     func(*ast.FuncDecl) bool
}

func NewParser() *Parser {
	return &Parser{
		Listing: &ResourceListing{
			Infos: Infomation{},
			Apis:  make([]*ApiRef, 0),
		},
		PackagesCache:    make(map[string]map[string]*ast.Package),
		TopLevelApis:     make(map[string]*ApiDeclaration),
		TypeDefinitions:  make(map[string]map[string]*ast.TypeSpec),
		PackagePathCache: make(map[string]string),
		PackageImports:   make(map[string]map[string]string),
	}
}

//Read web/main.go to get General info
func (parser *Parser) ParseGeneralApiInfo(mainApiFile string) {

	fileSet := token.NewFileSet()
	fileTree, err := goparser.ParseFile(fileSet, mainApiFile, nil, goparser.ParseComments)
	if err != nil {
		log.Fatalf("Can not parse general API information: %v\n", err)
	}

	parser.Listing.SwaggerVersion = SwaggerVersion
	if fileTree.Comments != nil {
		for _, comment := range fileTree.Comments {
			for _, commentLine := range strings.Split(comment.Text(), "\n") {
				if strings.HasPrefix(commentLine, "@APIVersion") {
					parser.Listing.ApiVersion = strings.TrimSpace(commentLine[len("@APIVersion"):])
				} else if strings.HasPrefix(commentLine, "@Title") {
					parser.Listing.Infos.Title = strings.TrimSpace(commentLine[len("@Title"):])
				} else if strings.HasPrefix(commentLine, "@Description") {
					parser.Listing.Infos.Description = strings.TrimSpace(commentLine[len("@Description"):])
				} else if strings.HasPrefix(commentLine, "@TermsOfServiceUrl") {
					parser.Listing.Infos.TermsOfServiceUrl = strings.TrimSpace(commentLine[len("@TermsOfServiceUrl"):])
				} else if strings.HasPrefix(commentLine, "@Contact") {
					parser.Listing.Infos.Contact = strings.TrimSpace(commentLine[len("@Contact"):])
				} else if strings.HasPrefix(commentLine, "@License") {
					parser.Listing.Infos.License = strings.TrimSpace(commentLine[len("@License"):])
				} else if strings.HasPrefix(commentLine, "@LicenseUrl") {
					parser.Listing.Infos.LicenseUrl = strings.TrimSpace(commentLine[len("@LicenseUrl"):])
				}
			}
		}
	}
}

func (parser *Parser) GetResourceListingJson() []byte {
	json, err := json.MarshalIndent(parser.Listing, "", "    ")
	if err != nil {
		log.Fatalf("Can not serialise ResourceListing to JSON: %v\n", err)
	}
	return json
}

func (parser *Parser) GetApiDescriptionJson() []byte {
	json, err := json.MarshalIndent(parser.TopLevelApis, "", "    ")
	if err != nil {
		log.Fatalf("Can not serialise []ApiDescription to JSON: %v\n", err)
	}
	return json
}

func (parser *Parser) CheckRealPackagePath(packagePath string) string {
	packagePath = strings.Trim(packagePath, "\"")

	if cachedResult, ok := parser.PackagePathCache[packagePath]; ok {
		return cachedResult
	}

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		log.Fatalf("Please, set $GOPATH environment variable\n")
	}

	pkgRealpath := ""
	gopathsList := filepath.SplitList(gopath)
	for _, path := range gopathsList {
		if evalutedPath, err := filepath.EvalSymlinks(filepath.Join(path, "src", packagePath)); err == nil {
			if _, err := os.Stat(evalutedPath); err == nil {
				pkgRealpath = evalutedPath
				break
			}
		}
	}
	if pkgRealpath == "" {
		goroot := filepath.Clean(runtime.GOROOT())
		if goroot == "" {
			log.Fatalf("Please, set $GOROOT environment variable\n")
		}
		if evalutedPath, err := filepath.EvalSymlinks(filepath.Join(goroot, "src", "pkg", packagePath)); err == nil {
			if _, err := os.Stat(evalutedPath); err == nil {
				pkgRealpath = evalutedPath
			}
		}
	}
	parser.PackagePathCache[packagePath] = pkgRealpath
	return pkgRealpath
}

func (parser *Parser) GetRealPackagePath(packagePath string) string {
	pkgRealpath := parser.CheckRealPackagePath(packagePath)
	if pkgRealpath == "" {
		log.Fatalf("Can not find package %s \n", packagePath)
	}

	return pkgRealpath
}

func (parser *Parser) GetPackageAst(packagePath string) map[string]*ast.Package {
	//log.Printf("Parse %s package\n", packagePath)
	if cache, ok := parser.PackagesCache[packagePath]; ok {
		return cache
	} else {
		fileSet := token.NewFileSet()

		astPackages, err := goparser.ParseDir(fileSet, packagePath, ParserFileFilter, goparser.ParseComments)
		if err != nil {
			log.Fatalf("Parse of %s pkg cause error: %s\n", packagePath, err)
		}
		parser.PackagesCache[packagePath] = astPackages
		return astPackages
	}
}

func (parser *Parser) AddOperation(op *Operation) {
	path := []string{}
	for _, pathPart := range strings.Split(op.Path, "/") {
		if pathPart = strings.TrimSpace(pathPart); pathPart != "" {
			path = append(path, pathPart)
		}
	}

	api, ok := parser.TopLevelApis[path[0]]
	if !ok {
		api = NewApiDeclaration()

		api.ApiVersion = parser.Listing.ApiVersion
		api.SwaggerVersion = SwaggerVersion
		api.ResourcePath = "/" + path[0]
		api.BasePath = parser.BasePath

		parser.TopLevelApis[path[0]] = api

		apiRef := &ApiRef{
			Path: api.ResourcePath,
		}
		parser.Listing.Apis = append(parser.Listing.Apis, apiRef)
	}

	api.AddOperation(op)
}

//TypeDefinitions
func (parser *Parser) ParseTypeDefinitions(packageName string) {
	parser.CurrentPackage = packageName
	pkgRealPath := parser.GetRealPackagePath(packageName)
	//	log.Printf("Parse type definition of %#v\n", packageName)

	if _, ok := parser.TypeDefinitions[pkgRealPath]; !ok {
		parser.TypeDefinitions[pkgRealPath] = make(map[string]*ast.TypeSpec)
	}

	astPackages := parser.GetPackageAst(pkgRealPath)
	for _, astPackage := range astPackages {
		for _, astFile := range astPackage.Files {
			for _, astDeclaration := range astFile.Decls {
				if generalDeclaration, ok := astDeclaration.(*ast.GenDecl); ok && generalDeclaration.Tok == token.TYPE {
					for _, astSpec := range generalDeclaration.Specs {
						if typeSpec, ok := astSpec.(*ast.TypeSpec); ok {
							parser.TypeDefinitions[pkgRealPath][typeSpec.Name.String()] = typeSpec
						}
					}
				}
			}
		}
	}

	//log.Fatalf("Type definition parsed %#v\n", parser.ParseImportStatements(packageName))

	for importedPackage, _ := range parser.ParseImportStatements(packageName) {
		//log.Printf("Import: %v, %v\n", importedPackage, v)
		parser.ParseTypeDefinitions(importedPackage)
	}
}

func (parser *Parser) ParseImportStatements(packageName string) map[string]bool {

	parser.CurrentPackage = packageName
	pkgRealPath := parser.GetRealPackagePath(packageName)

	imports := make(map[string]bool)
	astPackages := parser.GetPackageAst(pkgRealPath)

	parser.PackageImports[pkgRealPath] = make(map[string]string)
	for _, astPackage := range astPackages {
		for _, astFile := range astPackage.Files {
			for _, astImport := range astFile.Imports {
				importedPackageName := strings.Trim(astImport.Path.Value, "\"")
				if !IsIgnoredPackage(importedPackageName) {
					realPath := parser.GetRealPackagePath(importedPackageName)
					//log.Printf("path: %#v, original path: %#v", realPath, astImport.Path.Value)
					if _, ok := parser.TypeDefinitions[realPath]; !ok {
						imports[importedPackageName] = true
						//log.Printf("Parse %s, Add new import definition:%s\n", packageName, astImport.Path.Value)
					}

					importPath := strings.Split(importedPackageName, "/")
					parser.PackageImports[pkgRealPath][importPath[len(importPath)-1]] = importedPackageName
				}
			}
		}
	}
	return imports
}

func (parser *Parser) GetModelDefinition(model string, packageName string) *ast.TypeSpec {
	pkgRealPath := parser.CheckRealPackagePath(packageName)
	if pkgRealPath == "" {
		return nil
	}

	packageModels, ok := parser.TypeDefinitions[pkgRealPath]
	if !ok {
		return nil
	}
	astTypeSpec, _ := packageModels[model]
	return astTypeSpec
}

func (parser *Parser) FindModelDefinition(modelName string, currentPackage string) (*ast.TypeSpec, string) {
	var model *ast.TypeSpec
	var modelPackage string

	modelNameParts := strings.Split(modelName, ".")

	//if no dot in name - it can be only model from current package
	if len(modelNameParts) == 1 {
		modelPackage = currentPackage
		if model = parser.GetModelDefinition(modelName, currentPackage); model == nil {
			log.Fatalf("Can not find definition of %s model. Current package %s", modelName, currentPackage)
		}
	} else {
		//first try to assume what name is absolute
		absolutePackageName := strings.Join(modelNameParts[:len(modelNameParts)-1], "/")
		modelNameFromPath := modelNameParts[len(modelNameParts)-1]

		modelPackage = absolutePackageName
		if model = parser.GetModelDefinition(modelNameFromPath, absolutePackageName); model == nil {

			//can not get model by absolute name.
			if len(modelNameParts) > 2 {
				log.Fatalf("Can not find definition of %s model. Name looks like absolute, but model not found in %s package", modelNameFromPath, absolutePackageName)
			}

			// lets try to find it in imported packages
			pkgRealPath := parser.CheckRealPackagePath(currentPackage)
			if imports, ok := parser.PackageImports[pkgRealPath]; !ok {
				log.Fatalf("Can not find definition of %s model. Package %s dont import anything", modelNameFromPath, pkgRealPath)
			} else if relativePackage, ok := imports[modelNameParts[0]]; !ok {
				log.Fatalf("Package %s is not imported to %s, Imported: %#v\n", modelNameParts[0], currentPackage, imports)
			} else if model = parser.GetModelDefinition(modelNameFromPath, relativePackage); model == nil {
				log.Fatalf("Can not find definition of %s model in package %s", modelNameFromPath, relativePackage)
			} else {
				modelPackage = relativePackage
			}
		}
	}
	return model, modelPackage
}

func (parser *Parser) ParseApiDescription(packageName string) {
	parser.CurrentPackage = packageName
	pkgRealPath := parser.GetRealPackagePath(packageName)

	astPackages := parser.GetPackageAst(pkgRealPath)
	for _, astPackage := range astPackages {
		for _, astFile := range astPackage.Files {
			for _, astDescription := range astFile.Decls {
				switch astDeclaration := astDescription.(type) {
				case *ast.FuncDecl:
					if parser.IsController(astDeclaration) {
						operation := NewOperation(parser, packageName)
						if err := operation.ParseComment(astDeclaration.Doc); err != nil {
							if err != CommentIsEmptyError {
								log.Printf("Can not parse comment for function: %v, package: %v, got error: %v\n", astDeclaration.Name.String(), packageName, err)
							}
						} else {
							//log.Printf("Parsed comment: %#v\n", astDeclaration.Doc)
							parser.AddOperation(operation)
						}
					}
				}
			}
			for _, astComment := range astFile.Comments {
				for _, commentLine := range strings.Split(astComment.Text(), "\n") {
					parser.ParseSubApiDescription(commentLine)
				}
			}
		}
	}
}

// Parse sub api declaration
// @SubApi Very fance API [/fancy-api]
func (parser *Parser) ParseSubApiDescription(commentLine string) {
	if !strings.HasPrefix(commentLine, "@SubApi") {
		return
	} else {
		commentLine = strings.TrimSpace(commentLine[len("@SubApi"):])
	}
	re := regexp.MustCompile(`([\w\s]+)\[{1}([\w\_\-/]+)`)

	if matches := re.FindStringSubmatch(commentLine); len(matches) != 3 {
		log.Printf("Can not parse sub api description %s, skipped", commentLine)
	} else {
		for _, ref := range parser.Listing.Apis {
			if ref.Path == matches[2] {
				ref.Description = strings.TrimSpace(matches[1])
			}
		}
	}
}

func IsIgnoredPackage(packageName string) bool {
	return packageName == "C" || packageName == "appengine/cloudsql"
}

func ParserFileFilter(info os.FileInfo) bool {
	name := info.Name()
	return !info.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
}

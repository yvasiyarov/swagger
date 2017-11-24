package parser

import (
	"encoding/json"
	"fmt"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/yvasiyarov/swagger/utils"
)

var (
	log = logrus.WithField("pkg", "parser")
)

type Parser struct {
	APIPackages                       []string
	Listing                           *ResourceListing
	TopLevelApis                      map[string]*ApiDeclaration
	PackagesCache                     map[string]map[string]*ast.Package
	CurrentPackage                    string
	TypeDefinitions                   map[string]map[string]*ast.TypeSpec
	PackagePathCache                  map[string]string
	PackageImports                    map[string]map[string][]string
	BasePath, ControllerClass, Ignore string
	IsController                      func(*ast.FuncDecl, string) bool
	TypesImplementingMarshalInterface map[string]string

	VendoringPath    string
	DisableVendoring bool
	GoRoot           string
	GoPath           string
}

// It must return true if funcDeclaration is controller. We will try to parse only comments before controllers
func isController(funcDeclaration *ast.FuncDecl, controllerClass string) bool {
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

func NewParser(apiPackages, controllerClass, ignoreParam, vendoringPath string, disableVendoring bool) (*Parser, error) {
	gopath, goroot, err := utils.GetGoVars()
	if err != nil {
		return nil, err
	}

	packages := strings.Split(apiPackages, ",")

	return &Parser{
		APIPackages: packages,
		Listing: &ResourceListing{
			Infos: Infomation{},
			Apis:  make([]*ApiRef, 0),
		},
		IsController:     isController,
		ControllerClass:  controllerClass,
		Ignore:           ignoreParam,
		VendoringPath:    vendoringPath,
		DisableVendoring: disableVendoring,
		GoPath:           gopath,
		GoRoot:           goroot,
		PackagesCache:    make(map[string]map[string]*ast.Package),
		TopLevelApis:     make(map[string]*ApiDeclaration),
		TypeDefinitions:  make(map[string]map[string]*ast.TypeSpec),
		PackagePathCache: make(map[string]string),
		PackageImports:   make(map[string]map[string][]string),
		TypesImplementingMarshalInterface: map[string]string{
			"NullString":  "string",
			"NullInt64":   "int",
			"NullFloat64": "float",
			"NullBool":    "bool",
		},
	}, nil
}

func (parser *Parser) IsImplementMarshalInterface(typeName string) bool {
	_, ok := parser.TypesImplementingMarshalInterface[typeName]
	return ok
}

//Read web/main.go to get General info
func (parser *Parser) ParseGeneralApiInfo(mainApiFile string) {

	fileSet := token.NewFileSet()
	fileTree, err := goparser.ParseFile(fileSet, mainApiFile, nil, goparser.ParseComments)
	if err != nil {
		log.Fatalf("Can not parse general API information: %v\n", err)
	}

	parser.Listing.BasePath = "{{.}}"
	parser.Listing.SwaggerVersion = SwaggerVersion
	if fileTree.Comments != nil {
		for _, comment := range fileTree.Comments {
			for _, commentLine := range strings.Split(comment.Text(), "\n") {
				attribute := strings.ToLower(strings.Split(commentLine, " ")[0])
				switch attribute {
				case "@apiversion":
					parser.Listing.ApiVersion = strings.TrimSpace(commentLine[len(attribute):])
				case "@apititle":
					parser.Listing.Infos.Title = strings.TrimSpace(commentLine[len(attribute):])
				case "@apidescription":
					parser.Listing.Infos.Description = strings.TrimSpace(commentLine[len(attribute):])
				case "@termsofserviceurl":
					parser.Listing.Infos.TermsOfServiceUrl = strings.TrimSpace(commentLine[len(attribute):])
				case "@contact":
					parser.Listing.Infos.Contact = strings.TrimSpace(commentLine[len(attribute):])
				case "@licenseurl":
					parser.Listing.Infos.LicenseUrl = strings.TrimSpace(commentLine[len(attribute):])
				case "@license":
					parser.Listing.Infos.License = strings.TrimSpace(commentLine[len(attribute):])
				case "@basepath":
					parser.Listing.BasePath = strings.TrimSpace(commentLine[len(attribute):])
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

	// Hack vendoring of 'golang.org/x' by the standard library
	if strings.HasPrefix(packagePath, "golang_org/x/") {
		packagePath = filepath.Join("vendor", packagePath)
	}

	// Only check under vendor if it is not explicity disabled AND the package
	// in question is not an API package
	if !parser.DisableVendoring && !utils.StringSliceContains(parser.APIPackages, packagePath) {
		var searchPaths []string

		if parser.VendoringPath == "" {
			// CWD
			searchPaths = append(searchPaths, filepath.Join("vendor", packagePath))

			// $GOPATH/api_package/vendor/package_path
			for _, apiPkg := range parser.APIPackages {
				searchPaths = append(searchPaths, filepath.Join(parser.GoPath, "src", apiPkg, "vendor", packagePath))
			}
		} else {
			searchPaths = []string{
				fmt.Sprintf("%s/%s", parser.VendoringPath, packagePath),
			}
		}

		for _, path := range searchPaths {
			if evaluatedPath, err := filepath.EvalSymlinks(path); err == nil {
				if _, err := os.Stat(evaluatedPath); err == nil {
					log.Debugf("Found pkg '%v' in vendor dir (%v)", packagePath, evaluatedPath)
					parser.PackagePathCache[packagePath] = evaluatedPath
					return evaluatedPath
				}
			}
		}
	}

	// next, check GOPATH
	gopathsList := filepath.SplitList(parser.GoPath)
	for _, path := range gopathsList {
		if evaluatedPath, err := filepath.EvalSymlinks(filepath.Join(path, "src", packagePath)); err == nil {
			if _, err := os.Stat(evaluatedPath); err == nil {
				log.Debugf("Found pkg '%v' in GOPATH (%v)", packagePath, evaluatedPath)
				parser.PackagePathCache[packagePath] = evaluatedPath
				return evaluatedPath
			}
		}
	}

	// next, check GOROOT (/src)
	if evaluatedPath, err := filepath.EvalSymlinks(filepath.Join(parser.GoRoot, "src", packagePath)); err == nil {
		if _, err := os.Stat(evaluatedPath); err == nil {
			log.Debugf("Found pkg '%v' in GOROOT (%v)", packagePath, evaluatedPath)
			parser.PackagePathCache[packagePath] = evaluatedPath
			return evaluatedPath
		}
	}

	// next, check GOROOT (/src/pkg) (for golang < v1.4)
	if evaluatedPath, err := filepath.EvalSymlinks(filepath.Join(parser.GoRoot, "src", "pkg", packagePath)); err == nil {
		if _, err := os.Stat(evaluatedPath); err == nil {
			log.Debugf("Found pkg '%v' in GOROOT < v1.4 (%v)", packagePath, evaluatedPath)
			parser.PackagePathCache[packagePath] = evaluatedPath
			return evaluatedPath
		}
	}

	return ""
}

func (parser *Parser) GetRealPackagePath(packagePath string) string {
	pkgRealpath := parser.CheckRealPackagePath(packagePath)
	if pkgRealpath == "" {
		log.Fatalf("Can not find package %s", packagePath)
	}

	return pkgRealpath
}

func (parser *Parser) GetPackageAst(packagePath string) map[string]*ast.Package {
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

	resource := path[0]
	if op.ForceResource != "" {
		resource = op.ForceResource
	}

	api, ok := parser.TopLevelApis[resource]
	if !ok {
		api = NewApiDeclaration()

		api.ApiVersion = parser.Listing.ApiVersion
		api.SwaggerVersion = SwaggerVersion
		api.ResourcePath = "/" + resource
		api.BasePath = parser.Listing.BasePath

		parser.TopLevelApis[resource] = api
	}

	found := false
	for _, apiRef := range parser.Listing.Apis {
		if apiRef.Path == api.ResourcePath {
			found = true
		}
	}
	if !found {
		apiRef := &ApiRef{
			Path:        api.ResourcePath,
			Description: op.Summary,
		}
		parser.Listing.Apis = append(parser.Listing.Apis, apiRef)
	}

	api.AddOperation(op)
}

func (parser *Parser) ParseApi() {
	packages := parser.ScanPackages()

	for _, packageName := range packages {
		parser.ParseTypeDefinitions(packageName)
	}

	for _, packageName := range packages {
		parser.ParseApiDescription(packageName)
	}
}

func (parser *Parser) ScanPackages() []string {
	var res []string
	existsPackages := make(map[string]bool)

	for _, packageName := range parser.APIPackages {
		if v, ok := existsPackages[packageName]; !ok || v == false {
			// Add package
			existsPackages[packageName] = true
			res = append(res, packageName)

			// get it's real path
			pkgRealPath := parser.GetRealPackagePath(packageName)

			// Then walk
			var walker filepath.WalkFunc = func(path string, info os.FileInfo, err error) error {
				if !isIgnorablePath(path, info) {
					if idx := strings.Index(path, packageName); idx != -1 {
						pack := path[idx:]
						if v, ok := existsPackages[pack]; !ok || v == false {
							existsPackages[pack] = true
							res = append(res, pack)
						}
					}
				}
				return nil
			}

			filepath.Walk(pkgRealPath, walker)
		}
	}

	return res
}

func isIgnorablePath(path string, info os.FileInfo) bool {
	if !info.IsDir() {
		return true
	}

	ignorableContains := []string{
		"/_",      // avoid hidden dirs with initial "_"
		"/vendor", // anything that 'looks' like a vendor dir
		"/.git",   // anything that 'looks' like a git dir
	}

	for _, i := range ignorableContains {
		if strings.Contains(path, i) {
			return true
		}
	}

	return false
}

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

	parser.PackageImports[pkgRealPath] = make(map[string][]string)
	for _, astPackage := range astPackages {
		for _, astFile := range astPackage.Files {
			for _, astImport := range astFile.Imports {
				importedPackageName := strings.Trim(astImport.Path.Value, "\"")
				if !parser.isIgnoredPackage(importedPackageName) {
					realPath := parser.GetRealPackagePath(importedPackageName)
					//log.Printf("path: %#v, original path: %#v", realPath, astImport.Path.Value)
					if _, ok := parser.TypeDefinitions[realPath]; !ok {
						imports[importedPackageName] = true
						//log.Printf("Parse %s, Add new import definition:%s\n", packageName, astImport.Path.Value)
					}

					var importedPackageAlias string
					if astImport.Name != nil && astImport.Name.Name != "." && astImport.Name.Name != "_" {
						importedPackageAlias = astImport.Name.Name
					} else {
						importPath := strings.Split(importedPackageName, "/")
						importedPackageAlias = importPath[len(importPath)-1]
					}

					isExists := false
					for _, v := range parser.PackageImports[pkgRealPath][importedPackageAlias] {
						if v == importedPackageName {
							isExists = true
						}
					}

					if !isExists {
						parser.PackageImports[pkgRealPath][importedPackageAlias] = append(parser.PackageImports[pkgRealPath][importedPackageAlias], importedPackageName)
					}
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
			} else {
				var modelFound bool

				for _, packageName := range relativePackage {
					if model = parser.GetModelDefinition(modelNameFromPath, packageName); model != nil {
						modelPackage = packageName
						modelFound = true

						break
					}
				}

				if !modelFound {
					log.Fatalf("Can not find definition of %s model in package %s", modelNameFromPath, relativePackage)
				}
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
					if parser.IsController(astDeclaration, parser.ControllerClass) {
						operation := NewOperation(parser, packageName)
						if astDeclaration.Doc != nil && astDeclaration.Doc.List != nil {
							for _, comment := range astDeclaration.Doc.List {
								if err := operation.ParseComment(comment.Text); err != nil {
									log.Printf("Can not parse comment for function: %v, package: %v, got error: %v\n", astDeclaration.Name.String(), packageName, err)
								}
							}
						}
						if operation.Path != "" {
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
// @SubApi Very fancy API [/fancy-api]
func (parser *Parser) ParseSubApiDescription(commentLine string) {
	if !strings.HasPrefix(commentLine, "@SubApi") {
		return
	} else {
		commentLine = strings.TrimSpace(commentLine[len("@SubApi"):])
	}
	re := regexp.MustCompile(`([^\[]+)\[{1}([\w\_\-/]+)`)

	if matches := re.FindStringSubmatch(commentLine); len(matches) != 3 {
		log.Printf("Can not parse sub api description %s, skipped", commentLine)
	} else {
		found := false
		for _, ref := range parser.Listing.Apis {
			if ref.Path == matches[2] {
				found = true
				ref.Description = strings.TrimSpace(matches[1])
			}
		}
		if !found {
			subApi := &ApiRef{Path: matches[2],
				Description: strings.TrimSpace(matches[1]),
			}
			parser.Listing.Apis = append(parser.Listing.Apis, subApi)
		}
	}
}

func (parser *Parser) isIgnoredPackage(packageName string) bool {
	r, _ := regexp.Compile("appengine+")
	matched, err := regexp.MatchString(parser.Ignore, packageName)
	if err != nil {
		log.Fatalf("The -ignore argument is not a valid regular expression: %v\n", err)
	}
	return packageName == "C" || r.MatchString(packageName) || matched
}

func ParserFileFilter(info os.FileInfo) bool {
	name := info.Name()
	return !info.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
}

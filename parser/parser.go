package parser

import (
	"encoding/json"
	"fmt"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Parser struct {
	Listing        *ResourceListing
	TopLevelApis   map[string]*ApiDeclaration
	PackagesCache  map[string]map[string]*ast.Package
	CurrentPackage string
}

func NewParser() *Parser {
	return &Parser{
		Listing: &ResourceListing{
			Infos: Infomation{},
			Apis:  make([]*ApiRef, 0),
		},
		PackagesCache: make(map[string]map[string]*ast.Package),
		TopLevelApis:  make(map[string]*ApiDeclaration),
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

func (parser *Parser) IsController(funcDeclaration *ast.FuncDecl) bool {
	if funcDeclaration.Recv != nil && len(funcDeclaration.Recv.List) > 0 {
		if starExpression, ok := funcDeclaration.Recv.List[0].Type.(*ast.StarExpr); ok {
			receiverName := fmt.Sprint(starExpression.X)
			return strings.Index(receiverName, "Context") != -1
		}
	}
	return false
}
func GetRealPackagePath(packagePath string) string {
	packagePath = strings.Trim(packagePath, "\"")

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
	return pkgRealpath
}

func (parser *Parser) GetPackageAst(packagePath string) map[string]*ast.Package {
	log.Printf("Parse %s package\n", packagePath)
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
		parser.TopLevelApis[path[0]] = api

		apiRef := &ApiRef{
			Path: "/" + path[0],
		}
		parser.Listing.Apis = append(parser.Listing.Apis, apiRef)
	}

	api.AddOperation(op)
}

func (parser *Parser) ParseApiDescription(packageName string) {
	parser.CurrentPackage = packageName
	pkgRealPath := GetRealPackagePath(packageName)

	astPackages := parser.GetPackageAst(pkgRealPath)
	for _, astPackage := range astPackages {
		for _, astFile := range astPackage.Files {
			for _, astDescription := range astFile.Decls {
				switch astDeclaration := astDescription.(type) {
				case *ast.FuncDecl:
					if parser.IsController(astDeclaration) {
						operation := NewOperation(parser, packageName)
						if err := operation.ParseComment(astDeclaration.Doc, astDeclaration.Name.String()); err != nil {
							if err != CommentIsEmptyError {
								log.Printf("Can not parse comment for function: %v, package: %v, got error: %v\n", astDeclaration.Name.String(), packageName, err)
							}
						} else {
							parser.AddOperation(operation)
							//	log.Fatalf("Operation: %#v\n\n", operation)
							//				controllersList = append(controllersList, specDecl)
						}
					}
					/*
						case *ast.GenDecl:
							if specDecl.Tok.String() == "type" {
								for _, s := range specDecl.Specs {
									switch tp := s.(*ast.TypeSpec).Type.(type) {
									case *ast.StructType:
										_ = tp.Struct
										controllerComments[pkgpath+s.(*ast.TypeSpec).Name.String()] = specDecl.Doc.Text()
									}
								}
							}
					*/
				}
			}
		}
	}
}

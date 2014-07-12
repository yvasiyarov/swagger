package parser

import (
	"errors"
	"fmt"
	"go/ast"
	//	"log"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

const SwaggerVersion = "1.2"
const (
	ContentTypeJson  = "application/json"
	ContentTypeXml   = "application/xml"
	ContentTypePlain = "text/plain"
	ContentTypeHtml  = "text/html"
)

var CommentIsEmptyError = errors.New("Comment is empty")

type ResourceListing struct {
	ApiVersion     string `json:"apiVersion"`
	SwaggerVersion string `json:"swaggerVersion"`
	// BasePath       string `json:"basePath"`  obsolete in 1.1
	Apis  []*ApiRef  `json:"apis"`
	Infos Infomation `json:"info"`
}

type ApiRef struct {
	Path        string `json:"path"` // relative or absolute, must start with /
	Description string `json:"description"`
}

type Infomation struct {
	Title             string `json:"title,omitempty"`
	Description       string `json:"description,omitempty"`
	Contact           string `json:"contact,omitempty"`
	TermsOfServiceUrl string `json:"termsOfServiceUrl,omitempty"`
	License           string `json:"license,omitempty"`
	LicenseUrl        string `json:"licenseUrl,omitempty"`
}

// https://github.com/wordnik/swagger-core/blob/scala_2.10-1.3-RC3/schemas/api-declaration-schema.json
type ApiDeclaration struct {
	ApiVersion     string            `json:"apiVersion"`
	SwaggerVersion string            `json:"swaggerVersion"`
	BasePath       string            `json:"basePath"`
	ResourcePath   string            `json:"resourcePath"` // must start with /
	Consumes       []string          `json:"consumes,omitempty"`
	Produces       []string          `json:"produces,omitempty"`
	Apis           []*Api            `json:"apis,omitempty"`
	Models         map[string]*Model `json:"models,omitempty"`
}

func NewApiDeclaration() *ApiDeclaration {
	return &ApiDeclaration{
		Apis:     make([]*Api, 0),
		Models:   make(map[string]*Model),
		Consumes: make([]string, 0),
		Produces: make([]string, 0),
	}
}

func (api *ApiDeclaration) AddConsumedTypes(op *Operation) {
	for _, contextType := range op.Consumes {
		isExists := false
		for _, existType := range api.Consumes {
			if existType == contextType {
				isExists = true
				break
			}
		}
		if !isExists {
			api.Consumes = append(api.Consumes, contextType)
		}
	}
}

func (api *ApiDeclaration) AddProducesTypes(op *Operation) {
	for _, contextType := range op.Produces {
		isExists := false
		for _, existType := range api.Produces {
			if existType == contextType {
				isExists = true
				break
			}
		}
		if !isExists {
			api.Produces = append(api.Produces, contextType)
		}
	}
}
func (api *ApiDeclaration) AddModels(op *Operation) {
	//	log.Fatalf("OP models: %#v\n API: %#v\n", op.models, api.Models)
	for _, m := range op.models {
		if m != nil {
			if _, ok := api.Models[m.Id]; !ok {
				api.Models[m.Id] = m
			}
		}
	}
}

func (api *ApiDeclaration) AddSubApi(op *Operation) {

	isExists := false
	for _, subApi := range api.Apis {
		if subApi.Path == op.Path {
			isExists = true
			break
		}
	}
	if !isExists {
		subApi := NewApi()
		subApi.Path = op.Path
		api.Apis = append(api.Apis, subApi)
	}
}

func (api *ApiDeclaration) AddOperation(op *Operation) {
	api.AddProducesTypes(op)
	api.AddConsumedTypes(op)
	api.AddModels(op)
	api.AddSubApi(op)
}

type Api struct {
	Path        string       `json:"path"` // relative or absolute, must start with /
	Description string       `json:"description"`
	Operations  []*Operation `json:"operations,omitempty"`
}

func NewApi() *Api {
	return &Api{
		Operations: make([]*Operation, 0),
	}
}

type Operation struct {
	HttpMethod string `json:"httpMethod"`
	Nickname   string `json:"nickname"`
	Type       string `json:"type"` // in 1.1 = DataType
	// ResponseClass    string            `json:"responseClass"` obsolete in 1.2
	Summary          string            `json:"summary,omitempty"`
	Notes            string            `json:"notes,omitempty"`
	Parameters       []Parameter       `json:"parameters,omitempty"`
	ResponseMessages []ResponseMessage `json:"responseMessages,omitempty"` // optional
	Consumes         []string          `json:"consumes,omitempty"`
	Produces         []string          `json:"produces,omitempty"`
	Authorizations   []Authorization   `json:"authorizations,omitempty"`
	Protocols        []Protocol        `json:"protocols,omitempty"`
	Path             string            `json:`
	parser           *Parser
	models           []*Model
	packageName      string
}

func NewOperation(p *Parser, packageName string) *Operation {
	return &Operation{
		parser:      p,
		models:      make([]*Model, 0),
		packageName: packageName,
	}
}
func (operation *Operation) ParseComment(commentList *ast.CommentGroup) error {
	if commentList != nil && commentList.List != nil {
		for _, comment := range commentList.List {
			//log.Printf("Parse comemnt: %#v\n", c)
			commentLine := strings.TrimSpace(strings.TrimLeft(comment.Text, "//"))
			if strings.HasPrefix(commentLine, "@router") {
				if err := operation.ParseRouterComment(commentLine); err != nil {
					return err
				}
			} else if strings.HasPrefix(commentLine, "@Title") {
				operation.Nickname = strings.TrimSpace(commentLine[len("@Title"):])
			} else if strings.HasPrefix(commentLine, "@Description") {
				operation.Summary = strings.TrimSpace(commentLine[len("@Description"):])
			} else if strings.HasPrefix(commentLine, "@Success") {
				if err := operation.ParseSuccessComment(commentLine); err != nil {
					return err
				}
			} else if strings.HasPrefix(commentLine, "@Param") {
				if err := operation.ParseParamComment(commentLine); err != nil {
					return err
				}
			} else if strings.HasPrefix(commentLine, "@Failure") {
				if err := operation.ParseFailureComment(commentLine); err != nil {
					return err
				}
			} else if strings.HasPrefix(commentLine, "@Type") {
				operation.Type = strings.TrimSpace(commentLine[len("@Type"):])
			} else if strings.HasPrefix(commentLine, "@Accept") {
				if err := operation.ParseAcceptComment(commentLine); err != nil {
					return err
				}
			}
		}
	} else {
		return CommentIsEmptyError
	}
	return nil
}

// Parse params return []string of param properties
// @Param	queryText		form	      string	  true		        "The email for login"
// 			[param name]    [param type] [data type]  [is mandatory?]   [Comment]
func (operation *Operation) ParseParamComment(commentLine string) error {
	swaggerParameter := Parameter{}
	paramString := strings.TrimSpace(commentLine[len("@Param "):])

	parts := strings.Split(paramString, " ")
	notEmptyParts := make([]string, 0, len(parts))
	for _, paramPart := range parts {
		if paramPart != "" {
			notEmptyParts = append(notEmptyParts, paramPart)
		}
	}
	parts = notEmptyParts

	if len(parts) < 4 {
		return fmt.Errorf("Comments @Param at least should has 4 params")
	}
	swaggerParameter.Name = parts[0]
	swaggerParameter.ParamType = parts[1]
	swaggerParameter.Type = parts[2]
	swaggerParameter.DataType = parts[2]
	swaggerParameter.Required = strings.ToLower(parts[3]) == "true"
	swaggerParameter.Description = strings.Trim(strings.Join(parts[4:], " "), "\"")

	operation.Parameters = append(operation.Parameters, swaggerParameter)
	return nil
}

func (operation *Operation) ParseAcceptComment(commentLine string) error {
	accepts := strings.Split(strings.TrimSpace(strings.TrimSpace(commentLine[len("@Accept"):])), ",")
	for _, a := range accepts {
		switch a {
		case "json":
			operation.Consumes = append(operation.Consumes, ContentTypeJson)
			operation.Produces = append(operation.Produces, ContentTypeJson)
		case "xml":
			operation.Consumes = append(operation.Consumes, ContentTypeXml)
			operation.Produces = append(operation.Produces, ContentTypeXml)
		case "plain":
			operation.Consumes = append(operation.Consumes, ContentTypePlain)
			operation.Produces = append(operation.Produces, ContentTypePlain)
		case "html":
			operation.Consumes = append(operation.Consumes, ContentTypeHtml)
			operation.Produces = append(operation.Produces, ContentTypeHtml)
		}
	}
	return nil
}
func (operation *Operation) ParseFailureComment(commentLine string) error {
	response := ResponseMessage{}
	statement := strings.TrimSpace(commentLine[len("@Failure"):])

	var httpCode []rune
	var start bool
	for i, s := range statement {
		if unicode.IsSpace(s) {
			if start {
				response.Message = strings.TrimSpace(statement[i+1:])
				break
			} else {
				continue
			}
		}
		start = true
		httpCode = append(httpCode, s)
	}

	if code, err := strconv.Atoi(string(httpCode)); err != nil {
		return fmt.Errorf("Failure notation parse error: %v\n", err)
	} else {
		response.Code = code
	}
	operation.ResponseMessages = append(operation.ResponseMessages, response)
	return nil
}

func (operation *Operation) ParseRouterComment(commentLine string) error {
	elements := strings.TrimSpace(commentLine[len("@router"):])
	e1 := strings.SplitN(elements, " ", 2)
	if len(e1) < 1 {
		return errors.New("you should has router infomation")
	}
	operation.Path = e1[0]
	if len(e1) == 2 && e1[1] != "" {
		e1 = strings.SplitN(e1[1], " ", 2)
		operation.HttpMethod = strings.ToUpper(strings.Trim(e1[0], "[]"))
	} else {
		operation.HttpMethod = "GET"
	}
	return nil
}

// @Success 200 {object} model.OrderRow
func (operation *Operation) ParseSuccessComment(commentLine string) error {
	sourceString := strings.TrimSpace(commentLine[len("@Success"):])

	parts := strings.Split(sourceString, " ")
	notEmptyParts := make([]string, 0, len(parts))
	for _, paramPart := range parts {
		if paramPart != "" {
			notEmptyParts = append(notEmptyParts, paramPart)
		}
	}
	parts = notEmptyParts

	response := ResponseMessage{}
	if code, err := strconv.Atoi(parts[0]); err != nil {
		return errors.New("Success http code must be int")
	} else {
		response.Code = code
	}

	if parts[1] == "{object}" {
		if len(parts) < 3 {
			return errors.New("Success annotation error: object type must be specified")
		}
		model := NewModel(operation.parser)
		modelName := parts[2]
		//		if !strings.HasPrefix(modelName, operation.packageName) {
		//			modelName = operation.packageName + "." + modelName
		//		}
		if err, innerModels := model.ParseModel(modelName, operation.parser.CurrentPackage); err != nil {
			return err
		} else {
			operation.models = append(operation.models, model)
			operation.models = append(operation.models, innerModels...)
		}
	} else {
		response.Message = parts[2]
	}

	operation.ResponseMessages = append(operation.ResponseMessages, response)
	return nil
}

type Protocol struct {
}

type ResponseMessage struct {
	Code          int    `json:"code"`
	Message       string `json:"message"`
	ResponseModel string `json:"responseModel"`
}

type Parameter struct {
	ParamType     string `json:"paramType"` // path,query,body,header,form
	Name          string `json:"name"`
	Description   string `json:"description"`
	DataType      string `json:"dataType"` // 1.2 needed?
	Type          string `json:"type"`     // integer
	Format        string `json:"format"`   // int64
	AllowMultiple bool   `json:"allowMultiple"`
	Required      bool   `json:"required"`
	Minimum       int    `json:"minimum"`
	Maximum       int    `json:"maximum"`
}

type ErrorResponse struct {
	Code   int    `json:"code"`
	Reason string `json:"reason"`
}

type Model struct {
	Id         string                    `json:"id"`
	Required   []string                  `json:"required,omitempty"`
	Properties map[string]*ModelProperty `json:"properties"`
	parser     *Parser
	context    *ModelContext
}
type ModelContext struct {
	fullPackageName string
	//TODO: import list
}

func NewModel(p *Parser) *Model {
	return &Model{
		parser:  p,
		context: &ModelContext{},
	}
}

// modelName is something like package.subpackage.SomeModel or just "subpackage.SomeModel"
func (m *Model) ParseModel(modelName string, currentPackage string) (error, []*Model) {
	//log.Printf("ParseModel: %s, CurrentPackage %s \n", modelName, currentPackage)

	astTypeSpec, modelPackage := m.parser.FindModelDefinition(modelName, currentPackage)

	//fullNameParts := strings.Split(fullModelName, ".")
	//modelName := fullNameParts[len(fullNameParts)-1]
	//m.context.fullPackageName = strings.Join(fullNameParts[:len(fullNameParts)-1], "/")

	//log.Printf("Model name: %s , %s \n", fullModelName, m.context.fullPackageName)
	modelNameParts := strings.Split(modelName, ".")
	m.Id = strings.Join(append(strings.Split(modelPackage, "/"), modelNameParts[len(modelNameParts)-1]), ".")

	var innerModelList []*Model
	if astStructType, ok := astTypeSpec.Type.(*ast.StructType); ok {
		m.ParseFieldList(astStructType.Fields.List, modelPackage)
		usedTypes := make(map[string]bool)

		for _, property := range m.Properties {
			typeName := strings.Trim(property.Type, "[]")
			if property.IsBasicType(typeName) {
				continue
			}

			usedTypes[typeName] = true
		}

		//log.Printf("Before parse inner model list: %#v\n (%s)", usedTypes, modelName)
		innerModelList = make([]*Model, len(usedTypes))

		for typeName, _ := range usedTypes {
			typeModel := NewModel(m.parser)
			if err, typeInnerModels := typeModel.ParseModel(typeName, modelPackage); err != nil {
				//log.Printf("Parse Inner Model error %#v \n", err)
				return err, nil
			} else {
				//log.Printf("Inner model %v parsed, parsing %s \n", typeName, modelName)

				innerModelList = append(innerModelList, typeModel)
				innerModelList = append(innerModelList, typeInnerModels...)
			}
		}
		//log.Printf("After parse inner model list: %#v\n (%s)", usedTypes, modelName)
		//log.Fatalf("Inner model list: %#v\n", innerModelList)

	}

	//log.Printf("ParseModel finished %s \n", modelName)
	return nil, innerModelList
}

func (m *Model) ParseFieldList(fieldList []*ast.Field, modelPackage string) {
	if fieldList == nil {
		return
	}
	//log.Printf("ParseFieldList\n")

	m.Properties = make(map[string]*ModelProperty)
	for _, field := range fieldList {
		m.ParseModelProperty(field, modelPackage)
	}
}

func (m *Model) ParseModelProperty(field *ast.Field, modelPackage string) {
	var name string
	var innerModel *Model

	property := NewModelProperty()
	//log.Printf("field: %#v", field)
	property.Type = property.GetTypeAsString(field.Type)

	if len(field.Names) == 0 {
		//name is not specified, so struct is "embeded" in our model
		if astSelectorExpr, ok := field.Type.(*ast.SelectorExpr); ok {
			astTypeIdent, _ := astSelectorExpr.X.(*ast.Ident)

			name = astTypeIdent.Name + "." + strings.TrimPrefix(astSelectorExpr.Sel.Name, "*")

			innerModel = NewModel(m.parser)
			//log.Printf("Try to parse embeded type %s \n", name)
			//log.Fatalf("DEBUG: field: %#v\n, selector.X: %#v\n selector.Sel: %#v\n", field, astSelectorExpr.X, astSelectorExpr.Sel)
			innerModel.ParseModel(name, modelPackage)

			for innerFieldName, innerField := range innerModel.Properties {
				m.Properties[innerFieldName] = innerField
			}
			m.Required = append(m.Required, innerModel.Required...)
		}
	} else {
		name = field.Names[0].Name
	}

	//log.Printf("ParseModelProperty: %s, CurrentPackage %s, type: %s \n", name, modelPackage, property.Type)
	//Analyse struct fields annotations
	if field.Tag != nil {
		structTag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
		if tag := structTag.Get("json"); tag != "" {
			name = tag
		}
		if thriftTag := structTag.Get("thrift"); thriftTag != "" {
			tags := strings.Split(thriftTag, ",")
			if tags[0] != "" {
				name = tags[0]
			}
		}
		if required := structTag.Get("required"); required != "" {
			m.Required = append(m.Required, name)
		}
		if desc := structTag.Get("description"); desc != "" {
			property.Description = desc
		}
	}
	m.Properties[name] = property
}

type ModelProperty struct {
	Type        string            `json:"type"`
	Description string            `json:"description"`
	Items       map[string]string `json:"items,omitempty"`
	Format      string            `json:"format"`
}

func NewModelProperty() *ModelProperty {
	return &ModelProperty{}
}

// refer to builtin.go
var basicTypes = map[string]bool{
	"bool":       true,
	"uint":       true,
	"uint8":      true,
	"uint16":     true,
	"uint32":     true,
	"uint64":     true,
	"int":        true,
	"int8":       true,
	"int16":      true,
	"int32":      true,
	"int64":      true,
	"float32":    true,
	"float64":    true,
	"string":     true,
	"complex64":  true,
	"complex128": true,
	"byte":       true,
	"rune":       true,
	"uintptr":    true,
}

func (p *ModelProperty) IsBasicType(typeName string) bool {
	_, ok := basicTypes[typeName]
	return ok || strings.Contains(typeName, "interface")
}

func (p *ModelProperty) GetTypeAsString(fieldType interface{}) string {
	var realType string
	if astArrayType, ok := fieldType.(*ast.ArrayType); ok {
		//		log.Printf("arrayType: %#v\n", astArrayType)
		realType = fmt.Sprintf("[]%v", p.GetTypeAsString(astArrayType.Elt))
	} else if _, ok := fieldType.(*ast.InterfaceType); ok {
		realType = "interface"
	} else {
		if astStarExpr, ok := fieldType.(*ast.StarExpr); ok {
			realType = fmt.Sprint(astStarExpr.X)
			//			log.Printf("Get type as string (star expression)! %#v, type: %s\n", astStarExpr.X, fmt.Sprint(astStarExpr.X))
		} else if astSelectorExpr, ok := fieldType.(*ast.SelectorExpr); ok {
			packageNameIdent, _ := astSelectorExpr.X.(*ast.Ident)
			realType = packageNameIdent.Name + "." + astSelectorExpr.Sel.Name

			//			log.Printf("Get type as string(selector expression)! X: %#v , Sel: %#v, type %s\n", astSelectorExpr.X, astSelectorExpr.Sel, realType)
		} else {
			//			log.Printf("Get type as string(no star expression)! %#v , type: %s\n", fieldType, fmt.Sprint(fieldType))
			realType = fmt.Sprint(fieldType)
		}
	}
	return realType
}

// https://github.com/wordnik/swagger-core/wiki/authorizations
type Authorization struct {
	LocalOAuth OAuth  `json:"local-oauth"`
	ApiKey     ApiKey `json:"apiKey"`
}

// https://github.com/wordnik/swagger-core/wiki/authorizations
type OAuth struct {
	Type       string               `json:"type"`   // e.g. oauth2
	Scopes     []string             `json:"scopes"` // e.g. PUBLIC
	GrantTypes map[string]GrantType `json:"grantTypes"`
}

// https://github.com/wordnik/swagger-core/wiki/authorizations
type GrantType struct {
	LoginEndpoint        Endpoint `json:"loginEndpoint"`
	TokenName            string   `json:"tokenName"` // e.g. access_code
	TokenRequestEndpoint Endpoint `json:"tokenRequestEndpoint"`
	TokenEndpoint        Endpoint `json:"tokenEndpoint"`
}

// https://github.com/wordnik/swagger-core/wiki/authorizations
type Endpoint struct {
	Url              string `json:"url"`
	ClientIdName     string `json:"clientIdName"`
	ClientSecretName string `json:"clientSecretName"`
	TokenName        string `json:"tokenName"`
}

// https://github.com/wordnik/swagger-core/wiki/authorizations
type ApiKey struct {
	Type   string `json:"type"`   // e.g. apiKey
	PassAs string `json:"passAs"` // e.g. header
}

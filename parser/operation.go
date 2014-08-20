package parser

import (
	"errors"
	"fmt"
	//"go/ast"
	"regexp"
	"strconv"
	"strings"
)

type Operation struct {
	HttpMethod       string            `json:"httpMethod"`
	Nickname         string            `json:"nickname"`
	Type             string            `json:"type"`
	Items            OperationItems    `json:"items,omitempty"`
	Summary          string            `json:"summary,omitempty"`
	Notes            string            `json:"notes,omitempty"`
	Parameters       []Parameter       `json:"parameters,omitempty"`
	ResponseMessages []ResponseMessage `json:"responseMessages,omitempty"`
	Consumes         []string          `json:"-"`
	Produces         []string          `json:"produces,omitempty"`
	Authorizations   []Authorization   `json:"authorizations,omitempty"`
	Protocols        []Protocol        `json:"protocols,omitempty"`
	Path             string            `json:"-"`
	ForceResource    string
	parser           *Parser
	Models           []*Model `json:"-"`
	packageName      string
}
type OperationItems struct {
	Ref  string `json:"$ref,omitempty"`
	Type string `json:"type,omitempty"`
}

func NewOperation(p *Parser, packageName string) *Operation {
	return &Operation{
		parser:      p,
		Models:      make([]*Model, 0),
		packageName: packageName,
	}
}

func (operation *Operation) SetItemsType(itemsType string) {
	operation.Items = OperationItems{}
	if IsBasicType(itemsType) {
		operation.Items.Type = itemsType
	} else {
		operation.Items.Ref = itemsType
	}
}

func (operation *Operation) ParseComment(comment string) error {
	commentLine := strings.TrimSpace(strings.TrimLeft(comment, "//"))
	if strings.HasPrefix(commentLine, "@Router") {
		if err := operation.ParseRouterComment(commentLine); err != nil {
			return err
		}
	} else if strings.HasPrefix(commentLine, "@Resource") {
		resource := strings.TrimSpace(commentLine[len("@Resource"):])
		if resource[0:1] == "/" {
			resource = resource[1:]
		}
		operation.ForceResource = resource
	} else if strings.HasPrefix(commentLine, "@Title") {
		operation.Nickname = strings.TrimSpace(commentLine[len("@Title"):])
	} else if strings.HasPrefix(commentLine, "@Description") {
		operation.Summary = strings.TrimSpace(commentLine[len("@Description"):])
	} else if strings.HasPrefix(commentLine, "@Success") {
		sourceString := strings.TrimSpace(commentLine[len("@Success"):])
		if err := operation.ParseResponseComment(sourceString); err != nil {
			return err
		}
	} else if strings.HasPrefix(commentLine, "@Param") {
		if err := operation.ParseParamComment(commentLine); err != nil {
			return err
		}
	} else if strings.HasPrefix(commentLine, "@Failure") {
		sourceString := strings.TrimSpace(commentLine[len("@Failure"):])
		if err := operation.ParseResponseComment(sourceString); err != nil {
			return err
		}
	} else if strings.HasPrefix(commentLine, "@Accept") {
		if err := operation.ParseAcceptComment(commentLine); err != nil {
			return err
		}
	}

	operation.Models = operation.getUniqueModels()

	return nil
}

func (operation *Operation) getUniqueModels() []*Model {

	uniqueModels := make([]*Model, 0, len(operation.Models))
	modelIds := map[string]bool{}

	for _, model := range operation.Models {
		if _, exists := modelIds[model.Id]; exists {
			continue
		}
		uniqueModels = append(uniqueModels, model)
		modelIds[model.Id] = true
	}

	return uniqueModels
}

// Parse params return []string of param properties
// @Param	queryText		form	      string	  true		        "The email for login"
// 			[param name]    [param type] [data type]  [is mandatory?]   [Comment]
func (operation *Operation) ParseParamComment(commentLine string) error {
	swaggerParameter := Parameter{}
	paramString := strings.TrimSpace(commentLine[len("@Param "):])

	re := regexp.MustCompile(`([\w]+)[\s]+([\w]+)[\s]+([\w.]+)[\s]+([\w]+)[\s]+"([^"]+)"`)

	if matches := re.FindStringSubmatch(paramString); len(matches) != 6 {
		return fmt.Errorf("Can not parse param comment \"%s\", skipped.", paramString)
	} else {
		//TODO: if type is not simple, then add to Models[]
		swaggerParameter.Name = matches[1]
		swaggerParameter.ParamType = matches[2]
		swaggerParameter.Type = matches[3]
		swaggerParameter.DataType = matches[3]
		swaggerParameter.Required = strings.ToLower(matches[4]) == "true"
		swaggerParameter.Description = matches[5]

		operation.Parameters = append(operation.Parameters, swaggerParameter)
	}

	return nil
}

// @Accept  json
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

// @Router /customer/get-wishlist/{wishlist_id} [get]
func (operation *Operation) ParseRouterComment(commentLine string) error {
	sourceString := strings.TrimSpace(commentLine[len("@Router"):])

	re := regexp.MustCompile(`([\w\.\/\-{}]+)[^\[]+\[([^\]]+)`)
	var matches []string

	if matches = re.FindStringSubmatch(sourceString); len(matches) != 3 {
		return fmt.Errorf("Can not parse router comment \"%s\", skipped.", commentLine)
	}

	operation.Path = matches[1]
	operation.HttpMethod = strings.ToUpper(matches[2])
	return nil
}

// @Success 200 {object} model.OrderRow "Error message, if code != 200"
func (operation *Operation) ParseResponseComment(commentLine string) error {
	re := regexp.MustCompile(`([\d]+)[\s]+([\w\{\}]+)[\s]+([\w\-\.\/]+)[^"]*(.*)?`)
	var matches []string

	if matches = re.FindStringSubmatch(commentLine); len(matches) != 5 {
		return fmt.Errorf("Can not parse response comment \"%s\", skipped.", commentLine)
	}

	response := ResponseMessage{}
	if code, err := strconv.Atoi(matches[1]); err != nil {
		return errors.New("Success http code must be int")
	} else {
		response.Code = code
	}
	response.Message = strings.Trim(matches[4], "\"")

	typeName := ""
	if matches[3] == "error" {
		typeName = "string"
	} else if IsBasicType(matches[3]) {
		typeName = matches[3]
	} else {
		model := NewModel(operation.parser)
		response.ResponseModel = matches[3]
		if err, innerModels := model.ParseModel(response.ResponseModel, operation.parser.CurrentPackage); err != nil {
			return err
		} else {
			typeName = model.Id

			operation.Models = append(operation.Models, model)
			operation.Models = append(operation.Models, innerModels...)
		}
	}

	response.ResponseModel = typeName
	if response.Code == 200 {
		if matches[2] == "{array}" {
			operation.SetItemsType(typeName)
			operation.Type = "array"
		} else {
			operation.Type = typeName
		}
	}

	operation.ResponseMessages = append(operation.ResponseMessages, response)
	return nil
}

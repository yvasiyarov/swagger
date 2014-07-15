package parser

import (
	"errors"
	"fmt"
	"go/ast"
	"strconv"
	"strings"
	"unicode"
)

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

	if operation.Path == "" {
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
		response.ResponseModel = parts[2]
		if err, innerModels := model.ParseModel(response.ResponseModel, operation.parser.CurrentPackage); err != nil {
			return err
		} else {
			response.ResponseModel = model.Id

			operation.models = append(operation.models, model)
			operation.models = append(operation.models, innerModels...)
		}
	} else {
		response.Message = parts[2]
	}

	operation.ResponseMessages = append(operation.ResponseMessages, response)
	return nil
}

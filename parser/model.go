package parser

import (
	"fmt"
	"go/ast"
	"reflect"
	"strings"
)

type Model struct {
	Id         string                    `json:"id"`
	Required   []string                  `json:"required,omitempty"`
	Properties map[string]*ModelProperty `json:"properties"`
	parser     *Parser
}

func NewModel(p *Parser) *Model {
	return &Model{
		parser: p,
	}
}

// modelName is something like package.subpackage.SomeModel or just "subpackage.SomeModel"
func (m *Model) ParseModel(modelName string, currentPackage string) (error, []*Model) {
	//log.Printf("Before parse model |%s|, package: |%s|\n", modelName, currentPackage)

	astTypeSpec, modelPackage := m.parser.FindModelDefinition(modelName, currentPackage)

	modelNameParts := strings.Split(modelName, ".")
	m.Id = strings.Join(append(strings.Split(modelPackage, "/"), modelNameParts[len(modelNameParts)-1]), ".")

	var innerModelList []*Model
	if astStructType, ok := astTypeSpec.Type.(*ast.StructType); ok {
		m.ParseFieldList(astStructType.Fields.List, modelPackage)
		usedTypes := make(map[string]bool)

		for _, property := range m.Properties {
			typeName := property.Type
			if typeName == "array" {
				if property.Items.Type != "" {
					typeName = property.Items.Type
				} else {
					typeName = property.Items.Ref
				}
			}
			if IsBasicType(typeName) || m.parser.IsImplementMarshalInterface(typeName) {
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
				for _, property := range m.Properties {
					if property.Type == "array" {
						if property.Items.Ref == typeName {
							property.Items.Ref = typeModel.Id
						}
					} else {
						if property.Type == typeName {
							property.Type = typeModel.Id
						}
					}
				}
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

	typeAsString := property.GetTypeAsString(field.Type)
	//	log.Printf("Get type as string %s \n", typeAsString)

	if strings.HasPrefix(typeAsString, "[]") {
		property.Type = "array"
		property.SetItemType(typeAsString[2:])
	} else {
		property.Type = typeAsString
	}

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
			return
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
	Type        string             `json:"type"`
	Description string             `json:"description"`
	Items       ModelPropertyItems `json:"items,omitempty"`
	Format      string             `json:"format"`
}
type ModelPropertyItems struct {
	Ref  string `json:"$ref,omitempty"`
	Type string `json:"type,omitempty"`
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

func IsBasicType(typeName string) bool {
	_, ok := basicTypes[typeName]
	return ok || strings.Contains(typeName, "interface")
}

func (p *ModelProperty) SetItemType(itemType string) {
	p.Items = ModelPropertyItems{}
	if IsBasicType(itemType) {
		p.Items.Type = itemType
	} else {
		p.Items.Ref = itemType
	}
}
func (p *ModelProperty) GetTypeAsString(fieldType interface{}) string {
	var realType string
	if astArrayType, ok := fieldType.(*ast.ArrayType); ok {
		//		log.Printf("arrayType: %#v\n", astArrayType)
		realType = fmt.Sprintf("[]%v", p.GetTypeAsString(astArrayType.Elt))
	} else if astMapType, ok := fieldType.(*ast.MapType); ok {
		//		log.Printf("arrayType: %#v\n", astArrayType)
		realType = fmt.Sprintf("[]%v", p.GetTypeAsString(astMapType.Value))
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

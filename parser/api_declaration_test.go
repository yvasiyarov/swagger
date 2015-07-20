package parser_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/solher/swagger/parser"
	"testing"
)

type ApiDeclarationSuite struct {
	suite.Suite
	parser     *parser.Parser
	operation  *parser.Operation
	operation2 *parser.Operation
	operation3 *parser.Operation
}

func (suite *ApiDeclarationSuite) SetupSuite() {
	suite.parser = parser.NewParser()
	suite.operation = parser.NewOperation(suite.parser, "test")
	suite.operation2 = parser.NewOperation(suite.parser, "test")
	suite.operation3 = parser.NewOperation(suite.parser, "test")
}

func (suite *ApiDeclarationSuite) TestNewApi() {
	assert.NotNil(suite.T(), parser.NewApiDeclaration(), "Can no create new api description instance")
}

func (suite *ApiDeclarationSuite) TestAddConsumedTypes() {
	api := parser.NewApiDeclaration()

	suite.operation.Consumes = append(suite.operation.Consumes, parser.ContentTypeXml)
	suite.operation2.Consumes = append(suite.operation2.Consumes, parser.ContentTypeXml)
	suite.operation3.Consumes = append(suite.operation3.Consumes, parser.ContentTypeJson)

	api.AddConsumedTypes(suite.operation)
	assert.Len(suite.T(), api.Consumes, 1, "Consumed type was not added")
	api.AddConsumedTypes(suite.operation2)
	assert.Len(suite.T(), api.Consumes, 1, "Consumed type should be unique")
	api.AddConsumedTypes(suite.operation3)
	assert.Len(suite.T(), api.Consumes, 2, "Second consumed type was not added")

	expected := []string{parser.ContentTypeXml, parser.ContentTypeJson}
	assert.Equal(suite.T(), api.Consumes, expected, "Consumed types not added correctly")
}

func (suite *ApiDeclarationSuite) TestAddProducesTypes() {
	api := parser.NewApiDeclaration()

	suite.operation.Produces = append(suite.operation.Produces, parser.ContentTypeXml)
	suite.operation2.Produces = append(suite.operation2.Produces, parser.ContentTypeXml)
	suite.operation3.Produces = append(suite.operation3.Produces, parser.ContentTypeJson)

	api.AddProducesTypes(suite.operation)
	assert.Len(suite.T(), api.Produces, 1, "Produced type was not added")
	api.AddProducesTypes(suite.operation2)
	assert.Len(suite.T(), api.Produces, 1, "Produced type should be unique")
	api.AddProducesTypes(suite.operation3)
	assert.Len(suite.T(), api.Produces, 2, "Second produced type was not added")

	expected := []string{parser.ContentTypeXml, parser.ContentTypeJson}
	assert.Equal(suite.T(), api.Produces, expected, "Produced types not added correctly")
}

func (suite *ApiDeclarationSuite) TestAddModel() {
	api := parser.NewApiDeclaration()

	m1 := parser.NewModel(suite.parser)
	m1.Id = "test.SuperStruct"

	m2 := parser.NewModel(suite.parser)
	m2.Id = "test.SuperStruct"

	m3 := parser.NewModel(suite.parser)
	m3.Id = "test.SuperStruct2"

	suite.operation.Models = append(suite.operation.Models, m1)
	suite.operation2.Models = append(suite.operation2.Models, m2)
	suite.operation3.Models = append(suite.operation3.Models, m3)

	api.AddModels(suite.operation)
	assert.Len(suite.T(), api.Models, 1, "Model was not added")
	api.AddModels(suite.operation2)
	assert.Len(suite.T(), api.Models, 1, "Model should be unique")
	api.AddModels(suite.operation3)
	assert.Len(suite.T(), api.Models, 2, "Second model was not added")

	expected := map[string]*parser.Model{
		m1.Id: m1,
		m3.Id: m3,
	}
	assert.Equal(suite.T(), api.Models, expected, "Models not added correctly")
}

func (suite *ApiDeclarationSuite) TestAddSubApi() {
	api := parser.NewApiDeclaration()

	suite.operation.Path = "/customer/get{id}"
	suite.operation2.Path = "/customer/get{id}"
	suite.operation3.Path = "/order/get"

	api.AddSubApi(suite.operation)
	assert.Len(suite.T(), api.Apis, 1, "Api was not added")
	api.AddSubApi(suite.operation2)
	assert.Len(suite.T(), api.Apis, 1, "Api path should be unique")
	api.AddSubApi(suite.operation3)
	assert.Len(suite.T(), api.Apis, 2, "Second Api was not added")
}

func (suite *ApiDeclarationSuite) TestAddOperation() {
	api := parser.NewApiDeclaration()

	m1 := parser.NewModel(suite.parser)
	m1.Id = "test.SuperStruct"

	m2 := parser.NewModel(suite.parser)
	m2.Id = "test.SuperStruct"

	m3 := parser.NewModel(suite.parser)
	m3.Id = "test.SuperStruct2"

	suite.operation.Models = append(suite.operation.Models, m1)
	suite.operation2.Models = append(suite.operation2.Models, m2)
	suite.operation3.Models = append(suite.operation3.Models, m3)

	suite.operation.Path = "/customer/get{id}"
	suite.operation2.Path = "/customer/get{id}"
	suite.operation3.Path = "/order/get"

	suite.operation.Produces = append(suite.operation.Produces, parser.ContentTypeXml)
	suite.operation2.Produces = append(suite.operation2.Produces, parser.ContentTypeXml)
	suite.operation3.Produces = append(suite.operation3.Produces, parser.ContentTypeJson)

	suite.operation.Consumes = append(suite.operation.Consumes, parser.ContentTypeXml)
	suite.operation2.Consumes = append(suite.operation2.Consumes, parser.ContentTypeXml)
	suite.operation3.Consumes = append(suite.operation3.Consumes, parser.ContentTypeJson)

	api.AddOperation(suite.operation)
	api.AddOperation(suite.operation2)
	api.AddOperation(suite.operation3)

	assert.Len(suite.T(), api.Apis, 2, "Second Api was not added")

	expected := map[string]*parser.Model{
		m1.Id: m1,
		m3.Id: m3,
	}
	assert.Equal(suite.T(), api.Models, expected, "Models not added correctly")

	assert.Len(suite.T(), api.Produces, 2, "Second produced type was not added")
	assert.Len(suite.T(), api.Consumes, 2, "Second consumed type was not added")
}

func TestApiDeclarationSuite(t *testing.T) {
	suite.Run(t, &ApiDeclarationSuite{})
}

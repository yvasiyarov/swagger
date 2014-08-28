package parser_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/yvasiyarov/swagger/parser"
	"go/ast"
	"strings"
	"testing"
)

type ModelSuite struct {
	suite.Suite
	parser          *parser.Parser
	knownModelNames map[string]bool
}

var initialisedParser *parser.Parser

const ExamplePackageName = "github.com/yvasiyarov/swagger/example"

func (suite *ModelSuite) SetupSuite() {
	if initialisedParser == nil {
		initialisedParser = parser.NewParser()
		initialisedParser.ParseTypeDefinitions(ExamplePackageName)
	}
	suite.parser = initialisedParser
	suite.knownModelNames = make(map[string]bool)
}

func (suite *ModelSuite) GetExampleModelDefinition(modelName string) *ast.TypeSpec {
	var typeSpec *ast.TypeSpec
	for packageAbsolutePath, definitions := range suite.parser.TypeDefinitions {
		if strings.HasSuffix(packageAbsolutePath, ExamplePackageName) {
			if spec, ok := definitions[modelName]; ok {
				typeSpec = spec
				break
			} else {
				suite.T().Fatalf("Can not find model %s definition", modelName)
			}
		}
	}
	return typeSpec
}

func (suite *ModelSuite) TestNewModel() {
	assert.NotNil(suite.T(), parser.NewModel(suite.parser), "Can not create new model instance")
}

func (suite *ModelSuite) TestTypeDefinitions() {
	assert.NotEmpty(suite.T(), suite.parser.TypeDefinitions, "Can not parse type definitions from example package")
	assert.NotNil(suite.T(), suite.GetExampleModelDefinition("InterfaceType"), "Can not parse type definitions from example package")
}

func (suite *ModelSuite) TestInterfaceType() {
	m := parser.NewModel(suite.parser)
	err, innerModels := m.ParseModel("InterfaceType", ExamplePackageName, suite.knownModelNames)
	assert.Nil(suite.T(), err, "Can not parse InterfaceType definition")
	assert.Nil(suite.T(), innerModels, "Can not parse InterfaceType definition")

	assert.True(suite.T(), strings.HasSuffix(m.Id, "InterfaceType"), "Can not parse InterfaceType definition")
	assert.Len(suite.T(), m.Required, 0, "Can not parse InterfaceType definition")
	assert.Len(suite.T(), m.Properties, 0, "Can not parse InterfaceType definition")
}

func (suite *ModelSuite) TestSimpleAlias() {
	m := parser.NewModel(suite.parser)
	err, innerModels := m.ParseModel("SimpleAlias", ExamplePackageName, suite.knownModelNames)
	assert.Nil(suite.T(), err, "Can not parse SimpleAlias definition")
	assert.Nil(suite.T(), innerModels, "Can not parse SimpleAlias definition")

	assert.True(suite.T(), strings.HasSuffix(m.Id, "SimpleAlias"), "Can not parse SimpleAlias definition")
	assert.Len(suite.T(), m.Required, 0, "Can not parse SimpleAlias definition")
	assert.Len(suite.T(), m.Properties, 0, "Can not parse SimpleAlias definition")
}

func (suite *ModelSuite) TestSimpleStructure() {
	m := parser.NewModel(suite.parser)
	err, innerModels := m.ParseModel("SimpleStructure", ExamplePackageName, suite.knownModelNames)
	assert.Nil(suite.T(), err, "Can not parse SimpleStructure definition")
	assert.Len(suite.T(), innerModels, 0, "Can not parse SimpleStructure definition")

	assert.True(suite.T(), strings.HasSuffix(m.Id, "SimpleStructure"), "Can not parse SimpleStructuredefinition")
	assert.Len(suite.T(), m.Required, 0, "Can not parse SimpleStructure definition")
	assert.Len(suite.T(), m.Properties, 2, "Can not parse SimpleStructure definition")
}

func (suite *ModelSuite) TestSimpleStructureWithAnnotations() {
	m := parser.NewModel(suite.parser)
	err, innerModels := m.ParseModel("SimpleStructureWithAnnotations", ExamplePackageName, suite.knownModelNames)
	assert.Nil(suite.T(), err, "Can not parse SimpleStructureWithAnnotations definition")
	assert.Len(suite.T(), innerModels, 0, "Can not parse SimpleStructureWithAnnotations definition")

	assert.True(suite.T(), strings.HasSuffix(m.Id, "SimpleStructureWithAnnotations"), "Can not parse SimpleStructureWithAnnotations")
	assert.Len(suite.T(), m.Required, 1, "Can not parse SimpleStructureWithAnnotations definition(%#v)", m.Properties)
	assert.Len(suite.T(), m.Properties, 2, "Can not parse SimpleStructureWithAnnotations definition")

	assert.Equal(suite.T(), m.Properties["id"].Type, "int", "Can not parse SimpleStructureWithAnnotations definition")
	assert.Equal(suite.T(), m.Properties["Name"].Type, "string", "Can not parse SimpleStructureWithAnnotations definition")
}

func (suite *ModelSuite) TestStructureWithSlice() {
	m := parser.NewModel(suite.parser)
	err, innerModels := m.ParseModel("StructureWithSlice", ExamplePackageName, suite.knownModelNames)
	assert.Nil(suite.T(), err, "Can not parse StructureWithSlice definition")
	assert.Len(suite.T(), innerModels, 0, "Can not parse StructureWithSlice definition")

	assert.True(suite.T(), strings.HasSuffix(m.Id, "StructureWithSlice"), "Can not parse StructureWithSlice")
	assert.Len(suite.T(), m.Required, 0, "Can not parse StructureWithSlice definition(%#v)", m.Properties)
	assert.Len(suite.T(), m.Properties, 2, "Can not parse StructureWithSlice definition")

	assert.Equal(suite.T(), m.Properties["Id"].Type, "int", "Can not parse StructureWithSlice definition")
	assert.Equal(suite.T(), m.Properties["Name"].Type, "array", "Can not parse StructureWithSlice definition")
	assert.Equal(suite.T(), m.Properties["Name"].Items.Type, "byte", "Can not parse StructureWithSlice definition")
}

func (suite *ModelSuite) TestStructureWithEmbededStructure() {
	m := parser.NewModel(suite.parser)
	err, innerModels := m.ParseModel("StructureWithEmbededStructure", ExamplePackageName, suite.knownModelNames)
	assert.Nil(suite.T(), err, "Can not parse StructureWithEmbededStructure definition")

	assert.Len(suite.T(), innerModels, 0, "Can not parse StructureWithEmbededStructure definition (%#v)", innerModels)

	assert.True(suite.T(), strings.HasSuffix(m.Id, "StructureWithEmbededStructure"), "Can not parse StructureWithEmbededStructure")
	assert.Len(suite.T(), m.Required, 0, "Can not parse StructureWithEmbededStructure definition(%#v)", m.Properties)
	assert.Len(suite.T(), m.Properties, 2, "Can not parse StructureWithEmbededStructure definition")

	assert.Equal(suite.T(), m.Properties["Id"].Type, "int", "Can not parse StructureWithEmbededStructure definition")
	assert.Equal(suite.T(), m.Properties["Name"].Type, "array", "Can not parse StructureWithEmbededStructure definition")
	assert.Equal(suite.T(), m.Properties["Name"].Items.Type, "byte", "Can not parse StructureWithEmbededStructure definition")
}

func (suite *ModelSuite) TestStructureWithEmbededPointer() {
	m := parser.NewModel(suite.parser)
	err, innerModels := m.ParseModel("StructureWithEmbededPointer", ExamplePackageName, suite.knownModelNames)
	assert.Nil(suite.T(), err, "Can not parse StructureWithEmbededPointer definition")

	assert.Len(suite.T(), innerModels, 0, "Can not parse StructureWithEmbededPointer definition (%#v)", innerModels)

	assert.True(suite.T(), strings.HasSuffix(m.Id, "StructureWithEmbededPointer"), "Can not parse StructureWithEmbededPointer")
	assert.Len(suite.T(), m.Required, 0, "Can not parse StructureWithEmbededPointer definition(%#v)", m.Properties)
	assert.Len(suite.T(), m.Properties, 2, "Can not parse StructureWithEmbededPointer definition")

	assert.Equal(suite.T(), m.Properties["Id"].Type, "int", "Can not parse StructureWithEmbededPointer definition")
	assert.Equal(suite.T(), m.Properties["Name"].Type, "array", "Can not parse StructureWithEmbededPointer definition")
	assert.Equal(suite.T(), m.Properties["Name"].Items.Type, "byte", "Can not parse StructureWithEmbededPointer definition")
}

//TODO:
//embeded structures from other packages
//arrays of arrays

func TestModelSuite(t *testing.T) {
	suite.Run(t, &ModelSuite{})
}

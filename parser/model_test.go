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
	parser *parser.Parser
}

var initialisedParser *parser.Parser

func (suite *ModelSuite) SetupSuite() {
	if initialisedParser == nil {
		initialisedParser = parser.NewParser()
		initialisedParser.ParseTypeDefinitions("github.com/yvasiyarov/swagger/example")
	}
	suite.parser = initialisedParser
}

func (suite *ModelSuite) GetExampleModelDefinition(modelName string) *ast.TypeSpec {
	var typeSpec *ast.TypeSpec
	for packageAbsolutePath, definitions := range suite.parser.TypeDefinitions {
		if strings.HasSuffix(packageAbsolutePath, "github.com/yvasiyarov/swagger/example") {
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

func (suite *ModelSuite) TestParserModel() {
	assert.NotEmpty(suite.T(), suite.parser.TypeDefinitions, "Can not parse type definitions from example package")
	assert.NotNil(suite.T(), suite.GetExampleModelDefinition("InterfaceType"), "Can not parse type definitions from example package")
}

func TestModelSuite(t *testing.T) {
	suite.Run(t, &ModelSuite{})
}

package parser_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/yvasiyarov/swagger/parser"
	"go/ast"
	//	"log"
	"os"
	"path"
	"strings"
	"testing"
)

type ParserSuite struct {
	suite.Suite
	parser *parser.Parser
}

// It must return true if funcDeclaration is controller. We will try to parse only comments before controllers
func IsController(funcDeclaration *ast.FuncDecl) bool {
	if funcDeclaration.Recv != nil && len(funcDeclaration.Recv.List) > 0 {
		if starExpression, ok := funcDeclaration.Recv.List[0].Type.(*ast.StarExpr); ok {
			receiverName := fmt.Sprint(starExpression.X)
			return strings.Index(receiverName, "Context") != -1
		}
	}
	return false
}

var initialisedParser2 *parser.Parser
var exampleBasePath = "http://127.0.0.1:3000/"

func (suite *ParserSuite) SetupSuite() {
	if initialisedParser2 == nil {
		initialisedParser2 = parser.NewParser()

		initialisedParser2.BasePath = exampleBasePath
		initialisedParser2.IsController = IsController

		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			suite.T().Fatalf("Please, set $GOPATH environment variable\n")
		}

		initialisedParser2.ParseGeneralApiInfo(path.Join(gopath, "src", "github.com/yvasiyarov/swagger/example/web/main.go"))
		initialisedParser2.ParseApi("github.com/yvasiyarov/swagger/example")
	}
	suite.parser = initialisedParser2
}

func (suite *ParserSuite) TestNewParser() {
	assert.NotNil(suite.T(), suite.parser, "Parser instance was not created")
}

func (suite *ParserSuite) TestTopLevelAPI() {
	assert.Len(suite.T(), suite.parser.TopLevelApis, 1, "Top level API not parsed")
	if topApi, ok := suite.parser.TopLevelApis["testapi"]; !ok {
		suite.T().Fatalf("Can not find top level API:%v", suite.parser.TopLevelApis)
	} else {
		assert.Equal(suite.T(), exampleBasePath, topApi.BasePath, "Base path not set correctly")
		assert.NotEmpty(suite.T(), topApi.ApiVersion, "API version not filled")
		assert.NotEmpty(suite.T(), topApi.SwaggerVersion, "Swagger version not filled")
		assert.Equal(suite.T(), topApi.ResourcePath, "/testapi", "Resource path invalid")

		expectedTypes := []string{parser.ContentTypeJson}
		assert.Equal(suite.T(), topApi.Produces, expectedTypes, "Produced types not added correctly")
		assert.Equal(suite.T(), topApi.Consumes, expectedTypes, "Consumed types not added correctly")

		suite.CheckSubApiList(topApi)
		suite.CheckModelList(topApi)
	}
}

func (suite *ParserSuite) CheckSubApiList(topApi *parser.ApiDeclaration) {
	assert.Len(suite.T(), topApi.Apis, 9, "Sub API was not parsed corectly")

	for _, subApi := range topApi.Apis {
		switch subApi.Path {
		case "/testapi/get-string-by-int/{some_id}":
			assert.Equal(suite.T(), subApi.Description, "get string by ID", "Description was not parsed properly")
			assert.Len(suite.T(), subApi.Operations, 1, "Operations not parsed correctly")

			suite.CheckGetStringByInt(subApi.Operations[0])
		case "/testapi/get-struct-by-int/{some_id}":
			assert.Equal(suite.T(), subApi.Description, "get struct by ID", "Description was not parsed properly")
			assert.Len(suite.T(), subApi.Operations, 1, "Operations not parsed correctly")

			//suite.CheckGetStringByInt(subApi.Operations[0])
		case "/testapi/get-struct2-by-int/{some_id}":
			assert.Equal(suite.T(), subApi.Description, "get struct2 by ID", "Description was not parsed properly")
			assert.Len(suite.T(), subApi.Operations, 1, "Operations not parsed correctly")
		case "/testapi/get-simple-array-by-string/{some_id}":
			assert.Equal(suite.T(), subApi.Description, "get simple array by ID", "Description was not parsed properly")
			assert.Len(suite.T(), subApi.Operations, 1, "Operations not parsed correctly")
		case "/testapi/get-struct-array-by-string/{some_id}":
			assert.Equal(suite.T(), subApi.Description, "get struct array by ID", "Description was not parsed properly")
			assert.Len(suite.T(), subApi.Operations, 1, "Operations not parsed correctly")
		case "/testapi/get-interface":
			assert.Equal(suite.T(), subApi.Description, "get interface", "Description was not parsed properly")
			assert.Len(suite.T(), subApi.Operations, 1, "Operations not parsed correctly")
		case "/testapi/get-simple-aliased":
			assert.Equal(suite.T(), subApi.Description, "get simple aliases", "Description was not parsed properly")
			assert.Len(suite.T(), subApi.Operations, 1, "Operations not parsed correctly")
		case "/testapi/get-array-of-interfaces":
			assert.Equal(suite.T(), subApi.Description, "get array of interfaces", "Description was not parsed properly")
			assert.Len(suite.T(), subApi.Operations, 1, "Operations not parsed correctly")
		case "/testapi/get-struct3":
			assert.Equal(suite.T(), subApi.Description, "get struct3", "Description was not parsed properly")
			assert.Len(suite.T(), subApi.Operations, 1, "Operations not parsed correctly")
		default:
			suite.T().Fatalf("Undefined sub API: %#v", subApi)
		}
	}
}

func (suite *ParserSuite) CheckGetStringByInt(op *parser.Operation) {
	assert.Equal(suite.T(), "GET", op.HttpMethod, "Http method not parsed")
	assert.Equal(suite.T(), "GetStringByInt", op.Nickname, "Nickname not parsed")
	assert.Equal(suite.T(), "string", op.Type, "Type not parsed")
	assert.Equal(suite.T(), "string", op.Type, "Type not parsed")

	assert.Equal(suite.T(), op.Path, "/testapi/get-string-by-int/{some_id}", "Resource path invalid")

	expectedTypes := []string{parser.ContentTypeJson}
	assert.Equal(suite.T(), op.Produces, expectedTypes, "Produced types not added correctly")
	assert.Equal(suite.T(), op.Consumes, expectedTypes, "Consumed types not added correctly")

	assert.Len(suite.T(), op.Parameters, 1, "Params not parsed")
	assert.Len(suite.T(), op.ResponseMessages, 3, "Response message not parsed")

	suite.T().Log("Models list:")
	for _, m := range op.Models {
		suite.T().Log(m)
	}

	assert.Len(suite.T(), op.Models, 1, "Models not parsed %#v", op.Models)
}

func (suite *ParserSuite) CheckModelList(topApi *parser.ApiDeclaration) {
	assert.Len(suite.T(), topApi.Models, 6, "Models was not parsed corectly")

	for _, model := range topApi.Models {
		switch model.Id {
		case "github.com.yvasiyarov.swagger.example.APIError":
			assert.Len(suite.T(), model.Properties, 2, "Model not parsed correctly")
		}
	}
}

func TestParserSuite(t *testing.T) {
	suite.Run(t, &ParserSuite{})
}

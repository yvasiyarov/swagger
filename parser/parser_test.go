package parser_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/yvasiyarov/swagger/parser"
	"go/ast"
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
	suite.parser = parser.NewParser()
}

func (suite *ParserSuite) TestNewParser() {
	assert.NotNil(suite.T(), suite.parser, "Parser instance was not created")
}

func TestParserSuite(t *testing.T) {
	suite.Run(t, &ParserSuite{})
}

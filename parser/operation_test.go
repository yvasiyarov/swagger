package parser_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/yvasiyarov/swagger/parser"
	"testing"
)

type OperationSuite struct {
	suite.Suite
	parser *parser.Parser
}

func (suite *OperationSuite) SetupSuite() {
	suite.parser = parser.NewParser()
}

func (suite *OperationSuite) TestNewApi() {
	assert.NotNil(suite.T(), parser.NewOperation(suite.parser, "test"), "Can no create new operation instance")
}

func TestOperationSuite(t *testing.T) {
	suite.Run(t, &OperationSuite{})
}

package parser_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/solher/swagger/parser"
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

func (suite *OperationSuite) TestSetItemsType() {
	op := parser.NewOperation(suite.parser, "test")
	op.SetItemsType("int")
	assert.Equal(suite.T(), op.Items.Type, "int", "Can no set item type to simple type")

	op2 := parser.NewOperation(suite.parser, "test")
	op2.SetItemsType("SomeType")
	assert.Equal(suite.T(), op2.Items.Ref, "SomeType", "Can no set item type to custom type")
}

func (suite *OperationSuite) TestParseAcceptComment() {
	op := parser.NewOperation(suite.parser, "test")
	err := op.ParseAcceptComment("@Accept json")
	assert.Nil(suite.T(), err, "can not parse accept comment")
	assert.Equal(suite.T(), op.Consumes, []string{parser.ContentTypeJson}, "Can no parse accept comment")
	assert.Equal(suite.T(), op.Produces, []string{parser.ContentTypeJson}, "Can no parse accept comment")

	op2 := parser.NewOperation(suite.parser, "test")
	err2 := op2.ParseAcceptComment("@Accept json,html,plain,xml")
	assert.Nil(suite.T(), err2, "Can not parse accept comment with multiple types")

	expected := []string{parser.ContentTypeJson, parser.ContentTypeHtml, parser.ContentTypePlain, parser.ContentTypeXml}
	assert.Equal(suite.T(), op2.Consumes, expected, "Can not parse accept comment with multiple types")
	assert.Equal(suite.T(), op2.Produces, expected, "Can not parse accept comment with multiple types")
}

func (suite *OperationSuite) TestParseRouterComment() {
	op := parser.NewOperation(suite.parser, "test")
	err := op.ParseRouterComment("@Router /customer/get-wishlist/ [get]")
	assert.Nil(suite.T(), err, "Can not parse router comment")
	assert.Equal(suite.T(), op.Path, "/customer/get-wishlist/", "Can not parse router comment")
	assert.Equal(suite.T(), op.HttpMethod, "GET", "Can not parse router comment")

	op2 := parser.NewOperation(suite.parser, "test")
	err2 := op2.ParseRouterComment("@Router /customer/get-wishlist/{id} [PoSt]")
	assert.Nil(suite.T(), err2, "Can not parse router comment")
	assert.Equal(suite.T(), op2.Path, "/customer/get-wishlist/{id}", "Can not parse router comment")
	assert.Equal(suite.T(), op2.HttpMethod, "POST", "Can not parse router comment")
}

func (suite *OperationSuite) TestParseParamComment() {
	op := parser.NewOperation(suite.parser, "test")
	err := op.ParseParamComment("@Param   order_nr     path    string  true	\"Order number\"")
	assert.Nil(suite.T(), err, "Can not parse param comment")
	assert.Len(suite.T(), op.Parameters, 1, "Can not parse param comment")

	assert.Equal(suite.T(), op.Parameters[0].Name, "order_nr", "Can not parse param comment")
	assert.Equal(suite.T(), op.Parameters[0].ParamType, "path", "Can not parse param comment")
	assert.Equal(suite.T(), op.Parameters[0].Type, "string", "Can not parse param comment")
	assert.Equal(suite.T(), op.Parameters[0].DataType, "string", "Can not parse param comment")
	assert.Equal(suite.T(), op.Parameters[0].Required, true, "Can not parse param comment")
	assert.Equal(suite.T(), op.Parameters[0].Description, "Order number", "Can not parse param comment")
}

func (suite *OperationSuite) TestParseResponseComment() {
	op := parser.NewOperation(suite.parser, "test")
	err := op.ParseResponseComment("200 {simple} string")
	assert.Nil(suite.T(), err, "Can not parse response comment")
	assert.Len(suite.T(), op.ResponseMessages, 1, "Can not parse response comment")

	assert.Equal(suite.T(), op.ResponseMessages[0].Code, 200, "Can not parse response comment")
	assert.Equal(suite.T(), op.ResponseMessages[0].Message, "", "Can not parse response comment")
	assert.Equal(suite.T(), op.Type, "string", "Can not parse response comment")

	op2 := parser.NewOperation(suite.parser, "test")
	err2 := op2.ParseResponseComment("400 {simple} string     \"Order ID must be specified\"")
	assert.Nil(suite.T(), err2, "Can not parse response comment")
	assert.Len(suite.T(), op2.ResponseMessages, 1, "Can not parse response comment")

	assert.Equal(suite.T(), op2.ResponseMessages[0].Code, 400, "Can not parse response comment")
	assert.Equal(suite.T(), op2.ResponseMessages[0].Message, "Order ID must be specified", "Can not parse response comment")
	assert.Equal(suite.T(), op2.Type, "", "Can not parse response comment")

	op3 := parser.NewOperation(suite.parser, "test")
	err3 := op3.ParseResponseComment("200 {array} string ")
	assert.Nil(suite.T(), err3, "Can not parse response comment")
	assert.Len(suite.T(), op3.ResponseMessages, 1, "Can not parse response comment")

	assert.Equal(suite.T(), op3.ResponseMessages[0].Code, 200, "Can not parse response comment")
	assert.Equal(suite.T(), op3.ResponseMessages[0].Message, "", "Can not parse response comment")
	assert.Equal(suite.T(), op3.Type, "array", "Can not parse response comment")
	assert.Equal(suite.T(), op3.Items.Type, "string", "Can not parse response comment")
}

func (suite *OperationSuite) TestParseComment() {
	operationComment := `
// @Title getOrderByNumber
// @Description Return order by order number
// @Accept  json
// @Param   order_nr     path    string  true	"Order number"
// @Success 200 {array}  int
// @Failure 400 {simple} string     "Order ID must be specified"
// @Router order/by-number/{order_nr} [get]
`
	op := parser.NewOperation(suite.parser, "test")
	for _, line := range strings.Split(operationComment, "\n") {
		if len(line) > 0 { // the initial newline in the test string causes a panic otherwise...
			err := op.ParseComment(line)
			assert.Nil(suite.T(), err, "Can not parse operation comment")
		}
	}

	assert.Equal(suite.T(), op.Consumes, []string{parser.ContentTypeJson}, "Can no parse operation comment")
	assert.Equal(suite.T(), op.Produces, []string{parser.ContentTypeJson}, "Can no parse operation comment")

	assert.Equal(suite.T(), op.Path, "order/by-number/{order_nr}", "Can not parse operation comment")
	assert.Equal(suite.T(), op.HttpMethod, "GET", "Can not parse operation comment")

	assert.Len(suite.T(), op.Parameters, 1, "Can not parse operation comment")
	assert.Equal(suite.T(), op.Parameters[0].Name, "order_nr", "Can not parse operation comment")
	assert.Equal(suite.T(), op.Parameters[0].ParamType, "path", "Can not parse operation comment")
	assert.Equal(suite.T(), op.Parameters[0].Type, "string", "Can not parse operation comment")
	assert.Equal(suite.T(), op.Parameters[0].DataType, "string", "Can not parse operation comment")
	assert.Equal(suite.T(), op.Parameters[0].Required, true, "Can not parse operation comment")
	assert.Equal(suite.T(), op.Parameters[0].Description, "Order number", "Can not parse operation comment")

	assert.Len(suite.T(), op.ResponseMessages, 2, "Can not parse operation comment")
	assert.Equal(suite.T(), op.ResponseMessages[0].Code, 200, "Can not parse operation comment")
	assert.Equal(suite.T(), op.ResponseMessages[0].Message, "", "Can not parse operation comment")
	assert.Equal(suite.T(), op.Type, "array", "Can not parse operation comment")
	assert.Equal(suite.T(), op.Items.Type, "int", "Can not parse operation comment")

	assert.Equal(suite.T(), op.ResponseMessages[1].Code, 400, "Can not parse operation comment")
	assert.Equal(suite.T(), op.ResponseMessages[1].Message, "Order ID must be specified", "Can not parse operation comment")
}

func TestOperationSuite(t *testing.T) {
	suite.Run(t, &OperationSuite{})
}

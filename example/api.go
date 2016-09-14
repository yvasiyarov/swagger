// @SubApi Test API [/testapi]
package example

import (
	"encoding/json"
	"fmt"

	"github.com/gocraft/web"
	"github.com/yvasiyarov/swagger/example/subpackage"
)

type Context struct {
	response interface{}
}

func (c *Context) WriteResponse(response interface{}) {
	c.response = response
}

// @Title GetStringByInt
// @Description get string by ID
// @Accept  json
// @Produce  json
// @Param   some_id     path    int     true        "Some ID"
// @Success 200 {object} string
// @Failure 400 {object} APIError "We need ID!!"
// @Failure 404 {object} APIError "Can not find ID"
// @Router /testapi/get-string-by-int/{some_id} [get]
func (c *Context) GetStringByInt(rw web.ResponseWriter, req *web.Request) {
	c.WriteResponse(fmt.Sprint("Some data for %s ID", req.PathParams["some_id"]))
}

// @Title GetStructByInt
// @Description get struct by ID
// @Accept  json
// @Produce  json
// @Param   some_id     path    int     true        "Some ID"
// @Param   offset     query    int     true        "Offset"
// @Param   limit      query    int     true        "Offset"
// @Success 200 {object} StructureWithEmbededStructure
// @Failure 400 {object} APIError "We need ID!!"
// @Failure 404 {object} APIError "Can not find ID"
// @Router /testapi/get-struct-by-int/{some_id} [get]
func (c *Context) GetStructByInt(rw web.ResponseWriter, req *web.Request) {
	c.WriteResponse(StructureWithEmbededStructure{})
}

// @Title GetStruct2ByInt
// @Description get struct2 by ID
// @Accept  json
// @Produce  json
// @Param   some_id     path    int     true        "Some ID"
// @Param   offset     query    int     true        "Offset"
// @Param   limit      query    int     true        "Offset"
// @Success 200 {object} StructureWithEmbededPointer
// @Failure 400 {object} APIError "We need ID!!"
// @Failure 404 {object} APIError "Can not find ID"
// @Router /testapi/get-struct2-by-int/{some_id} [get]
func (c *Context) GetStruct2ByInt(rw web.ResponseWriter, req *web.Request) {
	c.WriteResponse(StructureWithEmbededPointer{})
}

// @Title GetSimpleArrayByString
// @Description get simple array by ID
// @Accept  json
// @Produce  json
// @Param   some_id     path    string     true        "Some ID"
// @Param   offset     query    int     true        "Offset"
// @Param   limit      query    int     true        "Offset"
// @Success 200 {array} string
// @Failure 400 {object} APIError "We need ID!!"
// @Failure 404 {object} APIError "Can not find ID"
// @Router /testapi/get-simple-array-by-string/{some_id} [get]
func (c *Context) GetSimpleArrayByString(rw web.ResponseWriter, req *web.Request) {
	c.WriteResponse([]string{"one", "two", "three"})
}

// @Title GetStructArrayByString
// @Description get struct array by ID
// @Accept  json
// @Produce  json
// @Param   some_id     path    string     true        "Some ID"
// @Param   offset     query    int     true        "Offset"
// @Param   limit      query    int     true        "Offset"
// @Success 200 {array} SimpleStructureWithAnnotations
// @Failure 400 {object} APIError "We need ID!!"
// @Failure 404 {object} APIError "Can not find ID"
// @Router /testapi/get-struct-array-by-string/{some_id} [get]
func (c *Context) GetStructArrayByString(rw web.ResponseWriter, req *web.Request) {
	c.WriteResponse([]subpackage.SimpleStructure{
		subpackage.SimpleStructure{},
		subpackage.SimpleStructure{},
		subpackage.SimpleStructure{},
	})
}

// @Title GetInterface
// @Description get interface
// @Accept  json
// @Produce  json
// @Success 200 {object} InterfaceType
// @Failure 400 {object} APIError "We need ID!!"
// @Failure 404 {object} APIError "Can not find ID"
// @Router /testapi/get-interface [get]
func (c *Context) GetInterface(rw web.ResponseWriter, req *web.Request) {
	c.WriteResponse(InterfaceType("Some string"))
}

// @Title GetSimpleAliased
// @Description get simple aliases
// @Accept  json
// @Produce  json
// @Success 200 {object} SimpleAlias
// @Failure 400 {object} APIError "We need ID!!"
// @Failure 404 {object} APIError "Can not find ID"
// @Router /testapi/get-simple-aliased [get]
func (c *Context) GetSimpleAliased(rw web.ResponseWriter, req *web.Request) {
	c.WriteResponse("Some string")
}

// @Title GetArrayOfInterfaces
// @Description get array of interfaces
// @Accept  json
// @Produce  json
// @Success 200 {array} InterfaceType
// @Failure 400 {object} APIError "We need ID!!"
// @Failure 404 {object} APIError "Can not find ID"
// @Router /testapi/get-array-of-interfaces [get]
func (c *Context) GetArrayOfInterfaces(rw web.ResponseWriter, req *web.Request) {
	c.WriteResponse([]InterfaceType{"Some string", 123, "10"})
}

// @Title GetStruct3
// @Description get struct3
// @Accept  json
// @Produce  json
// @Success 200 {object} StructureWithSlice
// @Failure 400 {object} APIError "We need ID!!"
// @Failure 404 {object} APIError "Can not find ID"
// @Router /testapi/get-struct3 [get]
func (c *Context) GetStruct3(rw web.ResponseWriter, req *web.Request) {
	c.WriteResponse(StructureWithSlice{})
}

func InitRouter() *web.Router {
	router := web.New(Context{}).
		Middleware(web.LoggerMiddleware).
		Middleware(web.ShowErrorsMiddleware).
		Middleware(func(c *Context, rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
			resultJSON, _ := json.Marshal(c.response)
			rw.Write(resultJSON)
		}).
		Get("/testapi/get-string-by-int/{some_id}", (*Context).GetStringByInt).
		Get("/testapi/get-struct-by-int/{some_id}", (*Context).GetStructByInt).
		Get("/testapi/get-simple-array-by-string/{some_id}", (*Context).GetSimpleArrayByString).
		Get("/testapi/get-struct-array-by-string/{some_id}", (*Context).GetStructArrayByString).
		Get("/testapi/get-interface", (*Context).GetInterface).
		Get("/testapi/get-simple-aliased", (*Context).GetSimpleAliased).
		Get("/testapi/get-array-of-interfaces", (*Context).GetArrayOfInterfaces).
		Get("/testapi/get-struct3", (*Context).GetStruct3).
		Get("/testapi/get-struct2-by-int/{some_id}", (*Context).GetStruct2ByInt)

	return router
}

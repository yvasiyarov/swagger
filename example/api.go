// @SubApi Order management API [/orders]
package example

import (
	"fmt"
	"github.com/gocraft/web"
)

type Context struct {
}

// @Title GetStringByInt
// @Description get string by ID
// @Accept  json
// @Param   some_id     path    int     true        "Some ID"
// @Success 200 {object} string
// @Failure 400 {object} APIError "We need ID!!"
// @Failure 404 {object} APIError "Can not find ID"
// @router /testapi/get-string-by-int/{some_id} [get]
func (c *Context) GetStringByInt(rw web.ResponseWriter, req *web.Request) {
	fmt.Fprint(rw, "Some data")
}

func InitRouter() *web.Router {
	router := web.New(Context{}).
		Middleware(web.LoggerMiddleware).
		Middleware(web.ShowErrorsMiddleware).
		Get("/get-string-by-int/{some_id}", (*Context).GetStringByInt)

	return router
}

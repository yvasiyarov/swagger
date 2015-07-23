// @APIVersion 1.0.0
// @APITitle Swagger Example API
// @APIDescription Swagger Example API
// @Contact varyous@gmail.com
// @TermsOfServiceUrl http://yvasiyarov.com/
// @License BSD
// @LicenseUrl http://yvasiyarov.com/
package main

import (
	"github.com/yvasiyarov/swagger/example"
	"net/http"
)

func main() {
	router := example.InitRouter()
	http.ListenAndServe("localhost:3000", router)
}

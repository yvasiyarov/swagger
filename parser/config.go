package parser

import (
//"github.com/yvasiyarov/web"
//"net/http"
)

type Config struct {
	ApiVersion      string
	WebServicesUrl  string // url where the services are available, e.g. http://localhost:8080
	ApiPath         string // path where the JSON api is avaiable , e.g. /apidocs
	SwaggerPath     string // [optional] path where the swagger UI will be served, e.g. /swagger
	SwaggerFilePath string // [optional] location of folder containing Swagger HTML5 application index.html
}

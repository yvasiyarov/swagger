Swagger UI generator for Go
===========================

This is a utility for automatically generating API documentation from annotations in Go code. It generates the documentation as JSON, according to the [Swagger Spec](https://github.com/wordnik/swagger-spec), and then displays it using [Swagger UI](https://github.com/wordnik/swagger-ui).

This tool was inspired by [Beego](http://beego.me/docs/advantage/docs.md), and follows the same annotation standards set by Beego. The main difference between this tool and Beego is that this generator doesn't depend on the Beego framework. You can use any framework to implement your API (or don't use a framework at all). You just add declarative comments to your API controllers, then run this generator and your documentation is ready! For an example of what such documentation looks like when presented via Swagger UI, see the Swagger [pet store example](http://petstore.swagger.wordnik.com/).

Project Status
--------------
This project is in an early-adopter, "Alpha," state. If you find a bug and think you know what the fix should be, please send pull request. If you want to help, please see list of TODO's at the end of this README.


Declarative Comments Format
---------------------------

### 1. General API info

Use the following annotation comments to describe the API as a whole.
They should be placed in the "main" file of your application, above the "package" keyword.
The @-tags are not case sensitive, but it is recommended to use the casing as shown, to be consistent.
Each of these annotations take a single argument that is an unquoted string to the end of the line.
The purpose of each annotation should be self-explanatory.
They are all optional, although using at least @Title and @Description is highly recommended.

    // @APIVersion 1.0.0
    // @Title My Cool API
    // @Description My API usually works as expected. But sometimes its not true
    // @Contact api@contact.me
    // @TermsOfServiceUrl http://google.com/
    // @License BSD
    // @LicenseUrl http://opensource.org/licenses/BSD-2-Clause



### 2. Sub API Definitions (One per Resource)

The Swagger specification is a bit confusing in how it refers to your API (singular) having multiple APIs (plural). It assumes that your API is "resource" centric. That is, it assumes that the first segment of every URL path refers to a "resource" and that it is therefore desirable to group the APIs specifications by these resources -- each with it's own API. (The [pet store example](http://petstore.swagger.wordnik.com/) is grouped according to three resources: pet, user, and store; with all of the URLs beginning with /pet/, /user/, or /store/, respectively).

NOTE: This is problematic for an application that is microservices-centric, rather than resource-centric. (See TODO, below.)

The @SubApi annotation is an opportunity to define each resource.

    // @SubApi Order management API [/orders]
    // @SubApi Statistic gavering API [/cache-stats]

@SubAPI comment should also be placed above the "package" keyword of the "main" file of your application. You can declare several sub-API's, one after the other. The format of the SubApi annotation is simple:

    // @SubApi DESCRIPTION [URI]

URI must have leading slash. This description is not mandatory, but if you forget it, then you will have an ugly looking document. :-)


### API Operation

The most important annotation comments for swagger UI generation are these comments that describe an operation (GET, PUT, POST, etc.). These comments are placed within your controller source code. One full set of these comments is used for each operation. They are placed just above the code that handles that operation.

Please, refer to the following example when reviewing the notes, below:

    // @Title getOrdersByCustomer
    // @Description retrieves orders for given customer defined by customer ID
    // @Accept  json
    // @Param   customer_id     path    int     true        "Customer ID"
    // @Param   order_id        query   int     false        "Retrieve order with given ID only"
    // @Param   order_nr        query   string  false        "Retrieve order with given number only"
    // @Param   created_from    query   string  false        "Date-time string, MySQL format. If specified, API will retrieve orders that were created starting from created_from"
    // @Param   created_to      query   string  false        "Date-time string, MySQL format. If specified, API will retrieve orders that were created before created_to"
    // @Success 200 {array}  my_api.model.OrderRow
    // @Failure 400 {object} my_api.ErrorResponse    Customer ID must be specified
    // @router /orders/by-customer/{customer_id} [get]

Let's describe every line in details:
* The @Title provides a "nickname", in Swagger terms, to the operation. It is kind of an "alias" for this API operation. Only [A-Za-z0-9] characters are allowed. It's required, but only used internally. Swagger UI does not display it.
* @Description - A longer description for the operation. (An unquoted string to the end of line.)
* @Accept - One of json/xml/plain/html. Should be equal to "Accept" header of your API.
* @Param - Defines a parameter that is accepted by this API operation. This comment have the following format:
 @Param  param_name  transport_type  data_type  required  "description"
 * param_name  - name of the parameter.
 * transport_type  - define how this parameter should be passed to your API. Can be one of path/query/form/header/body
 * data_type  - type of parameter
 * required - can be true or false
 * description - parameter description. Must be quoted.
* @Success/@Failure - Use these annotations to define the possible responses by the API operation. The format is as follows:
 @Success http_response_code response_type response_data_type response_description
 * http_response_code 200 for success response, any other code for failure.
 * response_type - can be {object} or {array} -- depending on whether the operation returns a single JSON object, or an array of objects
 * response_data_type - data type of your response. Can be one of the Go built-in types, or your custom type. All interface types will be displayed just as "interface". (It's not possible to find out which type it will has at parsing time.)
 * response_description - optional. It usually only makes sense for error responses.
* @router - define route path, which should be used to call this API operation. It has following format:
 @router request_path [request_method]
 * reqiest_path which should be used by to make request to this API endpoint. It can include placeholders for parameters with transport type equal "path". Look at the example above.
 * request_method - just HTTP request method (get/post/etc..)


Quick Start Guide
-----------------
TODO: implement example API :-)
1. Add comments to your API
2. Change IsController function in generator.go. It should return true if provided function declaration is "controller"
3. Run generator API:
  `go run generator.go -apiPackage="my_cool_api" -mainApiFile="my_cool_api/web/main.go" -basePath="http://127.0.0.1:3000"`
    * apiPackage  - package with API controllers implementation
    * mainApiFile - main API file. We will look for "General API info" in this file
    * basePath    - Your API URL. Test requests will be sent to this URL
  Once generator willl finish, you will find one new file in current directory - "docs.go"
4. Run generated swagger UI:
 `go run web.go docs.go`
5. Enjoy it :-)


TODO
----
1. Write better documentation
2. Document used data structures and methods
3. Figure out a recommended approach for documenting the API of a microservice-centric system, as opposed to a resource-centric system. This might include: documenting best practices, changing this generator code to understand additional annotation types, and/or recommending changes to the Swagger specification.

Known Limitations
-----------------
* Interface types are not supported, because it's not possible to resolve them to actual implementations are parse-time. All interface values will be displayed just as "interface".
* Types that implement the Marshaler/Unmarshaler interface. Marshaling of this types will produce unpredictable JSON (at parse-time).



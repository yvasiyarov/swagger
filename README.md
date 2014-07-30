Swagger UI generator for Go   
===========================   

Swagger UI generator was inspired by http://beego.me/docs/advantage/docs.md. Main difference from Beego's implementation - this generator doesn't depend on Beego framework. You can use any framework for your API (or don't use frameworks at all). You just add declarative comments to API controlers, which you wanna export, run generator and [your documentation is ready!](http://petstore.swagger.wordnik.com/)        

PS   
You should consider quality of this package as "alpha". If you find bug - just send pull request. If you want to help - please, see list of TODO's at the end of documentation.   


Declarative comments format  
---------------------------

### 1. General API info  

This comments should be placed in "main" file of your application, before "package" keyword. Below is example of such comments:   

    // @APIVersion 1.0.0
    // @Title My Cool API
    // @Description My API usually works as expected. But sometimes its not true 
    // @Contact api@contact.me
    // @TermsOfServiceUrl http://google.com/
    // @License BSD
    // @LicenseUrl http://opensource.org/licenses/BSD-2-Clause

Purpose of this comments are obvious and most of them is not mandatory.  

### 2. Sub API definition   
Swagger specification is a bit confusing in part of using "API" term. At least two different entities is called just "API", and non of them is really correct :-). If you will open [demo API](http://petstore.swagger.wordnik.com/), you will see what all API controlers is grouped by first part of URI into groups - "Sub API's". You can use "@SubAPI" comment to provide description to this groups of API controllers:   

    // @SubApi Order management API [/orders]
    // @SubApi Statistic gavering API [/cache-stats]

@SubAPI comment should be placed before "package" keyword. You can declare several sub API's in one file. Format of this comment is simple:    

    // @SubApi DESCRIPTION [URI]

URI must have leading slash. This comment is not mandatory. If you will forget it - you will just have ugly-looking documentation :-)   

### API operation   
This comments is most important for swagger UI generation. Please, see example below:  

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

Lets describe every line in details:  
* @Title comment is "nickname" in swagger terms. Kind of "alias" for this API operation. Only [a-Z0-9] characters are allowed.  Its used only in automatic swagger client generators
* @Description - just operation description
* @Accept - one of json/xml/plain/html. Should be equal to "Accept" header of your API
* @Param - define parameter, accepted by this API operation. This comment have following format:
 @Param  param_name  transport_type  data_type  mandatory  "description"    
 * param_name  - name of the parameter.
 * transport_type  - define how this parameter should be passed to your API. Can be one of path/query/form/header/body
 * data_type  - type of parameter   
 * mandatory - can be true or false
 * description - parameter description. Must be quoted.  
* @Success/@Failure - define values, which can be returned by API call. Format is follow:  
 @Success http_response_code response_type response_data_type response_description  
 * http_response_code 200 for success response, any other code for failure. 
 * response_type - can be {object} or {array} - if your method return array of elements
 * response_data_type - data type of your response. Can be one of Go build in types, or your custom type. All interface types will be displayed just as "interface". Its not possible to find out which type it will has at parsing time.  
 * response_description - optional. Usually make sense only for error responses
* @router - define route path, which should be used to call this API operation. It has following format:  
 @router request_path [request_method]    
 * reqiest_path which should be used by to make request to this API endpoint. It can include placeholders for parameters with transport type equal "path". Look at the example above. 
 * request_method - just HTTP request method (get/post/etc..)
 

Quick start guide    
--------------
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

Known limitations
-------
* Interface types is not supported, because its not possible to understand to which actual value interface will reference. All interface values displayed just as "interface" 
* Types which implement Marshaler/Unmarshaler interface. Marshaling of this types will produce unpredicable(at parsing time) JSON



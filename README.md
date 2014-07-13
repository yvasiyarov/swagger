Swagger UI generator for Go   
===========================   

Swagger UI generator was inspired by http://beego.me/docs/advantage/docs.md. You should consider quality of this package as "pre-alpha"
If you find bug - just send pull request.

How its works   
--------------
1. You should add special comment blocks to your API. This comments will  describe each API call, parameters which it expect and result which it can return
2. Change IsController function in generator.go.  
3. We will parse this comments and create JSON description of your API:
  go run generator.go -apiPackage="my_cool_api" -mainApiFile="my_cool_api/web/main.go" -basePath="http://127.0.0.1:3000"
  * apiPackage  - package which implement API controllers
  * mainApiFile - main API file. We will look for "main comment block" in this file
  * basePath    - URL, where your API service is running. Web interface will send requests to that URL

  Once generator willl finish, you will find one new file in current directory - "docs.go"
4. Run generated swagger UI:
`go run web.go docs.go`
5. Enjoy it :-)

Main comment block
-------------------
Main comment block must be located before "package" keyword, and it consists from following lines:  

    // @APIVersion 1.0.0
    // @Title My Cool API
    // @Description My API usually works as expected. But sometimes its not true 
    // @Contact api@contact.me
    // @TermsOfServiceUrl http://google.com/
    // @License BSD
    // @LicenseUrl http://opensource.org/licenses/BSD-2-Clause

Sub API definition comment
--------------------------
Sub API definition comment block must be located before "package" keyword, and it consists from one line. You can declare several sub API in one file:

    // @SubApi Order management API [/orders]
    // @SubApi Statistic gavering API [/cache-stats]

API call definition comment
---------------------------
It should be located just before controller function and it have following format

    // @Title getOrdersByCustomer
    // @Description retrieves orders for given customer defined by customer ID
    // @Accept  json
    // @Param   customer_id     path    int     true        "Customer ID"
    // @Param   order_id        query   int     false        "Retrieve order with given ID only"
    // @Param   order_nr        query   string  false        "Retrieve order with given number only"
    // @Param   created_from    query   string  false        "Date-time string, MySQL format. If specified, API will retrieve orders that were created starting from created_from"
    // @Param   created_to      query   string  false        "Date-time string, MySQL format. If specified, API will retrieve orders that were created before created_to"
    // @Success 200 {object} lazada_api.model.OrderRow
    // @Failure 400 Customer ID must be specified
    // @Failure 404 Customer not found
    // @router /orders/by-customer/:customer_id:\\d+ [get]


TODO
----
1. Add unit tests.
2. Write better documentation
3. Write examples of API
4. Document used data structures and methods
5. Investigate how we should work with arrays/slices/maps
6. Add possibility to parse API, which consists of one and more package
7. Refactor comment parsing, I should use reg exp for this purposes

Known limitations
-------
* Interface types will never be supported because not possible to understand to which actual value interface will reference. All interface values displayed just as "interface" 
* Types which implement Marshaler/Unmarshaler interface. Marshaling of this types will produce unpredicable(at parsing time) JSON



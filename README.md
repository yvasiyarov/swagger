Swagger UI generator for Go   
===========================   

Swagger UI generator was inspired by http://beego.me/docs/advantage/docs.md. You should consider quality of this package as "pre-alpha"
If you find bug - just send pull request.

How its works   
--------------
1. You should add special comment blocks to your API. This comments will  describe each API call, parameters which it expect and result which it can return
2. We will parse this comments and create JSON description of your API:
  go run generator.go -apiPackage="my_cool_api" -mainApiFile="my_cool_api/web/main.go" -basePath="http://127.0.0.1:3000"
  * apiPackage  - package which implement API controllers
  * mainApiFile - main API file. We will look for "main comment block" in this file
  * basePath    - URL, where your API service is running. Web interface will send requests to that URL

  Once generator willl finish, you will find one new file in current directory - "docs.go"
3. Run generated swagger UI:
`go run web.go docs.go`

Main comment block:

It consists from following comments:  

    // @APIVersion 1.0.0
    // @Title My Cool API
    // @Description My API usually works as expected. But sometimes its not true 
    // @Contact api@contact.me
    // @TermsOfServiceUrl http://google.com/
    // @License BSD
    // @LicenseUrl http://opensource.org/licenses/BSD-2-Clause


TODO:
1. Add unit tests.
2. Write better documentation
3. Write examples of API
4. Document used data structures and methods
5. Investigate how we should work with arrays/slices/maps
6. Add possibility to parse API, which consists of one and more package
7. Add possibility to pass "controller filter function" as parameter 

PS
Interface types will never be supported:
interfaces: its not possible to understand to which actual value interface will be referenced. All interface values displayed just as "interface" 


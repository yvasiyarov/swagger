
![alt text]( https://s3.amazonaws.com/tw-chat/attach/579528d6e2f2c2aebfe7f957e4572ca0/1.png  "Logo Title Text 1")


## Swagger UI Generator for Go



### About

This is a utility for automatically generating API documentation from annotations in Go code. It generates the documentation as JSON, according to the [Swagger Spec](https://github.com/wordnik/swagger-spec), and then displays it using [Swagger UI](https://github.com/swagger-api/swagger-ui).

This tool was inspired by [Beego](http://beego.me/docs/advantage/docs.md), and follows the same annotation standards set by Beego.
The main difference between this tool and Beego is that this generator doesn't depend on the Beego framework. You can use any framework to implement your API (or don't use a framework at all). You just add declarative comments to your API controllers, then run this generator and your documentation is ready! For an example of what such documentation looks like when presented via Swagger UI, see the Swagger [pet store example](http://petstore.swagger.wordnik.com/).

This tool focuses on _documentation generation_ as opposed to _client_ generation. If that is all you need, it will be significantly easier to integrate this tool into your existing codebase/workflow as opposed to [goswagger](https://goswagger.io/). One significant advantage of this tool is that it allows you to easily reference objects that are outside of your package.

_This tool currently generates Swagger 1.x spec files -- there are plans to update the tool to support Swagger 2.x at some point._

### Quick Start Guide

1. Add comments to your API source code, [see Declarative Comments Format ](https://github.com/yvasiyarov/swagger/wiki/Declarative-Comments-Format)

2. Download Swagger for Go by using ```go get github.com/yvasiyarov/swagger```

3. Or, compile the Swagger generator from sources.
    `go install`

    This will create a binary in your $GOPATH/bin folder called swagger (Mac/Unix) or swagger.exe (Windows).

4. Run the Swagger generator.
    Make sure to specify the full package name and an optional entry point (if the entry point isn't `$pkg/main.go`).

    Example:
    
    ```
    $ pwd
    /Users/dselans/Code/go/src/github.com/yvasiyarov/swagger
    $ ./$GOPATH/bin/swagger -apiPackage="github.com/yvasiyarov/swagger/example" -mainApiFile=example/web/main.go -output=./API.md -format=markdown
    ```
    
### Command Line Flags
|  Switch  |  Description   |
|------------------|---------------------------|    
| **-apiPackage**  | Package with API controllers implementation |
| **-mainApiFile** | Main API file. This file is used for generating the "General API Info" bits. If `-mainApiFile` is not specified, then `$apiPackage/main.go` is assumed. | 
| **-format**      | One of `go\|swagger\|asciidoc\|markdown\|confluence`. Default is `-format="go"`. See [docs](https://github.com/yvasiyarov/swagger/wiki/Generate-Different-Formats). |
| **-output**     | Output specification. Default varies according to -format. See [docs](https://github.com/yvasiyarov/swagger/wiki/Generate-Different-Formats). |
| **controllerClass**  | Speed up parsing by specifying which receiver objects have the controller methods. The default is to search all methods. The argument can be a regular expression. For example, `-controllerClass="(Context|Controller)$"` means the receiver name must end in Context or Controller. |
| **contentsTable**     | Whether to generate Table of Contents; default: `true`. |
| **models**       | Generate 'Models' section; default `true`. |
| **vendoringPath** | Override default vendor directory (eg. `$CWD/vendor` and `$GOPATH/src/$apiPackage/vendor`) |
| **disableVendoring** | Disable vendor usage altogether | 
| **enableDebug** | Enable debug log output |

### Note on Swagger-UI

To run the generated swagger UI (assuming you used -format="go"), copy/move the generated docs.go file to a new folder under GOPATH/src. Also bring in the web.go-example file, renaming it to web.go. Then: `go run web.go docs.go`

### Additional Documentation

**Project Status** : [Alpha](https://github.com/yvasiyarov/swagger/wiki/Declarative-Comments-Format)

**Declarative Comments Format** : [Read more ](https://github.com/yvasiyarov/swagger/wiki/Declarative-Comments-Format)

**Technical Notes** : [Read More](https://github.com/yvasiyarov/swagger/wiki/Technical-Notes)

**Known Limitations** : [Read More](https://github.com/yvasiyarov/swagger/wiki/Known-Limitations)
    
 **Generating Different Format Docs**: [Read More](https://github.com/yvasiyarov/swagger/wiki/Generate-Different-Formats)

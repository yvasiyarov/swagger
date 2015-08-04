$(function () {

    var url = window.location.search.match(/url=([^&]+)/);
    if (url && url.length > 1) {
        url = decodeURIComponent(url[1]);
    } else {
        url = window.location.protocol + "//" + window.location.host + "/deskapi/";
    }
    window.swaggerUi = new SwaggerUi({

        url: url,
        dom_id: "swagger-ui-container",
        useJQuery: true,
        withCredentials: true,
        sorter: "alpha",
        operationsSorter: "alpha",
        showRequestHeaders: true,
        headers: {
            "Authorization": "Basic UUxrNG5ocENHYTJQbTJzRnlYZE46Zm9v"
        },
        supportHeaderParams: true,
        supportedSubmitMethods: ['get', 'post', 'put', 'delete'],
        onComplete: function (swaggerApi, swaggerUi) {
            
            $('pre code').each(function (i, e) {
                hljs.highlightBlock(e)
            });
         addApiKeyAuthorization();
        },
        onFailure: function (data) {
            log("Unable to Load SwaggerUI");
        },
        docExpansion: "none",
        apisSorter: "alpha",
        showRequestHeaders: true
    });

    
      function addApiKeyAuthorization(){
        var key = encodeURIComponent($('#input_apiKey')[0].value);
        if(key && key.trim() != "") {
            var apiKeyAuth = new SwaggerClient.ApiKeyAuthorization("api_key", key, "query");
            window.swaggerUi.api.clientAuthorizations.add("api_key", apiKeyAuth);
            log("added key " + key);
        }
      }
      $('#input_apiKey').change(addApiKeyAuthorization);
      // if you have an apiKey you would like to pre-populate on the page for demonstration purposes...
      /*
        var apiKey = "myApiKeyXXXX123456789";
        $('#input_apiKey').val(apiKey);
      */
      window.swaggerUi.load();
      function log() {
        if ('console' in window) {
          console.log.apply(console, arguments);
        }
      }
  });
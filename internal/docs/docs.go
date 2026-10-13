// Package docs serves the backend OpenAPI specification and a Swagger UI to explore it.
package docs

import (
	_ "embed"
	"net/http"

	"github.com/labstack/echo/v4"
)

//go:embed openapi.yaml
var openAPISpec []byte

const swaggerUIVersion = "5.17.14"

const swaggerUIPage = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Omnibus Backend API</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@` + swaggerUIVersion + `/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@` + swaggerUIVersion + `/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: "/docs/openapi.yaml",
      dom_id: "#swagger-ui",
      persistAuthorization: true,
      withCredentials: true
    });
  </script>
</body>
</html>`

// SpecHandler returns the openapi.yaml file.
func SpecHandler(c echo.Context) error {
	return c.Blob(http.StatusOK, "application/yaml; charset=utf-8", openAPISpec)
}

// UIHandler returns the Swagger UI page that loads /docs/openapi.yaml.
func UIHandler(c echo.Context) error {
	return c.HTML(http.StatusOK, swaggerUIPage)
}

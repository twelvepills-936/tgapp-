package swagger

import (
	"net/http"
	"strings"

	openapi "github.com/twelvepills-936/tgapp-/api"
)

const swaggerHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = () => {
      SwaggerUIBundle({ url: "/openapi.json", dom_id: "#swagger-ui" });
    };
  </script>
</body>
</html>
`

// Wrap registers GET / (info), and when swaggerEnabled: GET /openapi.json and /swagger/.
func Wrap(next http.Handler, swaggerEnabled bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if r.Method == http.MethodGet && p == "/" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			if swaggerEnabled {
				w.Write([]byte(`{"ok":true,"swagger":"/swagger/","openapi":"/openapi.json"}`))
			} else {
				w.Write([]byte(`{"ok":true,"swagger":null,"openapi":null,"hint":"set SWAGGER_ENABLED=true"}`))
			}
			return
		}
		if !swaggerEnabled {
			next.ServeHTTP(w, r)
			return
		}
		switch {
		case p == "/openapi.json" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(openapi.OpenAPISpec)
		case p == "/swagger" && r.Method == http.MethodGet:
			http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
		case (p == "/swagger/" || p == "/swagger/index.html") && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(swaggerHTML))
		case strings.HasPrefix(p, "/swagger/"):
			http.NotFound(w, r)
		default:
			next.ServeHTTP(w, r)
		}
	})
}

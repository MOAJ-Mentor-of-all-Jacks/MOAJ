package enterprise

import "net/http"

// Bridge is the core-to-enterprise boundary. Both current implementations are
// deliberately disabled; a tagged build is not a license or an entitlement.
type Bridge interface {
	Edition() string
	Enabled(feature string) bool
	RegisterRoutes(mux *http.ServeMux)
}

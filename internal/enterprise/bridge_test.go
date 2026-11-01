package enterprise

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPreviewAndCommunityGrantNoEnterpriseCapabilities(t *testing.T) {
	b := New()
	if b.Edition() != Edition() {
		t.Fatal("bridge edition does not match build")
	}
	for _, feature := range []string{"saml", "oidc", "scim", "pbac", "siem", "unknown"} {
		if b.Enabled(feature) {
			t.Fatalf("unimplemented feature %q was enabled", feature)
		}
	}
	mux := http.NewServeMux()
	b.RegisterRoutes(mux)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/enterprise", nil))
	if rr.Code != http.StatusNotFound {
		t.Fatal("disabled bridge registered an enterprise route")
	}
}

//go:build ee

package enterprise

import "net/http"

// Edition marks the compile-only enterprise bridge. No paid capabilities are
// registered or unlocked until licensing is implemented in Phase 9.
func Edition() string { return "enterprise-preview" }

type previewBridge struct{}

func New() Bridge                                   { return previewBridge{} }
func (previewBridge) Edition() string               { return Edition() }
func (previewBridge) Enabled(string) bool           { return false }
func (previewBridge) RegisterRoutes(*http.ServeMux) {}

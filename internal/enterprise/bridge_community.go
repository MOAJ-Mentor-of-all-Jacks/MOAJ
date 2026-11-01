//go:build !ee

package enterprise

import "net/http"

// Edition identifies this build. Community has no enterprise implementation.
func Edition() string { return "community" }

type communityBridge struct{}

func New() Bridge                                     { return communityBridge{} }
func (communityBridge) Edition() string               { return Edition() }
func (communityBridge) Enabled(string) bool           { return false }
func (communityBridge) RegisterRoutes(*http.ServeMux) {}

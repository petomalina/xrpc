package multiplexer

import "net/http"

// HTTPHandler automatically fulfills all requests that come to its presence. This is
// because all non-http requests to an http server should be either redirected or fulfilled
// by other components in the way.
func HTTPHandler(handler http.Handler, selectors ...Selector) Handler {
	filter := append([]Selector{}, selectors...)

	return func(w http.ResponseWriter, r *http.Request) bool {
		for _, f := range filter {
			if !f(r) {
				return false
			}
		}

		handler.ServeHTTP(w, r)
		return true
	}
}

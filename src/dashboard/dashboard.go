// small monitoring dashboard for live monitoring and data visualization
package dashboard

import (
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"

	"fwdns/src/cache"
	"fwdns/src/stats"
)

// Bake assets into the binary
//
//go:embed assets
var content embed.FS

// bundles the things dashboard will need to read from
type Server struct {
	stats *stats.Stats
	cache *cache.Cache
}

// constructor for server
func New(s *stats.Stats, c *cache.Cache) *Server {
	return &Server{stats: s, cache: c}
}

// returns HTTP routes
//
// GET /          -> dashboard HTML page
// GET /api/stats -> current statistics as JSON
func (srv *Server) Handler() http.Handler {
	// serve files from embeded assets directory
	assets, err := fs.Sub(content, "assets")
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/stats", srv.handleStats)
	mux.Handle("/", http.FileServer(http.FS(assets)))
	return mux
}

// returns the current statistics as JSON
func (srv *Server) handleStats(w http.ResponseWriter, _ *http.Request) {
	snap := srv.stats.Snapshot()
	snap.CacheSize = srv.cache.Len()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(snap)
}

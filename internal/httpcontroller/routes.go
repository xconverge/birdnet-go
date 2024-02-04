// httpcontroller/routes.go
package httpcontroller

import (
	"embed"
	"html/template"
	"io/fs"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Embed the assets and views directories.
var AssetsFs embed.FS
var ViewsFs embed.FS

// routeConfig defines the structure for each route.
type routeConfig struct {
	Path         string
	TemplateName string
	Title        string // New field for page title
}

// routes lists all the routes in the application.
var routes = []routeConfig{
	{Path: "/", TemplateName: "dashboard", Title: "Dashboard"},
	{Path: "/dashboard", TemplateName: "dashboard", Title: "Dashboard"},
	{Path: "/logs", TemplateName: "logs", Title: "Logs"},
	{Path: "/stats", TemplateName: "stats", Title: "Statistics"},
	{Path: "/settings", TemplateName: "settings", Title: "General Settings"},
}

// initRoutes initializes the routes for the server.
func (s *Server) initRoutes() {
	// Define function map for templates.
	funcMap := template.FuncMap{
		"even":            even,
		"calcWidth":       calcWidth,
		"heatmapColor":    heatmapColor,
		"title":           cases.Title(language.English).String,
		"confidence":      confidence,
		"confidenceColor": confidenceColor,
		"RenderContent":   s.RenderContent,
		"sub":             func(a, b int) int { return a - b },
		"add":             func(a, b int) int { return a + b },
	}

	// Parse templates from the embedded filesystem.
	tmpl, err := template.New("").Funcs(funcMap).ParseFS(ViewsFs, "views/*.html", "views/**/*.html")
	if err != nil {
		s.Echo.Logger.Fatal(err)
	}
	s.Echo.Renderer = &TemplateRenderer{templates: tmpl}

	// Set up routes from the configuration.
	for _, route := range routes {
		s.Echo.GET(route.Path, s.handleRequest)
	}

	// Set up static file serving for assets.
	assetsFS, err := fs.Sub(AssetsFs, "assets")
	if err != nil {
		s.Echo.Logger.Fatal(err)
	}
	customFileServer(s.Echo, assetsFS, "assets")

	// Other static routes.
	s.Echo.Static("/clips", "clips")
	s.Echo.Static("/spectrograms", "spectrograms")

	// Additional handlers.
	s.Echo.GET("/top-birds", s.topBirdsHandler)
	s.Echo.GET("/notes", s.GetAllNotes)
	s.Echo.GET("/last-detections", s.GetLastDetections)
	s.Echo.GET("/species-detections", s.speciesDetectionsHandler)
	s.Echo.GET("/search", s.searchHandler)

	s.Echo.POST("/update-settings", s.updateSettingsHandler)

	// Specific handler for settings route
	//s.Echo.GET("/settings", s.settingsHandler)
}

package domain

const DefaultLinesPerFile = 300

var ExcludeAttrsOptimized = map[string]bool{
	"@ads":       true,
	"@tracking":  true,
	"@analytics": true,
}

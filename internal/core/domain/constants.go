package domain

const (
	DefaultLinesPerFile = 300

	DefaultPartCount = 6
)

var ExcludeAttrsOptimized = map[string]bool{
	"@ads":       true,
	"@tracking":  true,
	"@analytics": true,
}

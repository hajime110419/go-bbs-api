package utils

import "html"

// Sanitize escapes potentially harmful characters from a string to prevent
// Cross-Site Scripting (XSS) attacks.
func Sanitize(s string) string {
	return html.EscapeString(s)
}

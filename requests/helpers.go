package requests

import "strings"

func getContentType(path string) string {
	if strings.HasSuffix(path, "css") {
		return "text/css"
	} else if strings.HasSuffix(path, "js") {
		return "application/javascript"
	} else if strings.HasSuffix(path, "html") {
		return "text/html"
	}
	return "text/plain"
}

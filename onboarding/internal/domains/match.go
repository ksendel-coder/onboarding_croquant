package domains

import (
	"net/url"
	"strings"
)

func MatchPath(pattern, path string) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return false
	}
	if pattern == "*" || pattern == "/*" {
		return true
	}

	if !strings.Contains(pattern, "?") {
		if i := strings.IndexByte(path, '?'); i >= 0 {
			path = path[:i]
		}
	}

	segments := strings.Split(pattern, "*")

	if len(segments) == 1 {
		return trimSlash(pattern) == trimSlash(path)
	}

	rest := path
	for i, seg := range segments {
		if seg == "" {
			continue
		}

		switch {
		case i == 0:
			if !strings.HasPrefix(rest, seg) {
				return false
			}
			rest = rest[len(seg):]

		case i == len(segments)-1:
			return strings.HasSuffix(rest, seg)

		default:
			idx := strings.Index(rest, seg)
			if idx < 0 {
				return false
			}
			rest = rest[idx+len(seg):]
		}
	}

	return true
}

func PathFromURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	if parsed.Scheme == "" && parsed.Host == "" {
		if raw == "" {
			return "/"
		}
		return raw
	}

	path := parsed.Path
	if path == "" {
		path = "/"
	}

	if parsed.RawQuery != "" {
		path += "?" + parsed.RawQuery
	}

	return path
}

func trimSlash(s string) string {
	if len(s) > 1 && strings.HasSuffix(s, "/") {
		return strings.TrimRight(s, "/")
	}
	return s
}

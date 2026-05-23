package output

import "net/url"

func displayURL(raw string, showURLs bool) string {
	if showURLs {
		return raw
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "..."
	}

	if u.Scheme == "" || u.Host == "" {
		return "..."
	}

	u.User = nil
	u.RawQuery = ""
	u.Fragment = ""

	if u.Path != "" && u.Path != "/" {
		u.Path = "/..."
	}

	return u.String()
}

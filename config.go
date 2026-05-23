package main

import "os/user"

type config struct {
	Theme  string  `toml:"theme"`
	Name   *string `toml:"name"`
	Footer string  `toml:"footer"`
}

// presenterName returns the configured name, falling back to the current OS
// user's full name (or login name) only when no name key is present. An explicit
// empty name in the header is honoured as "no name".
func presenterName(c config) string {
	if c.Name != nil {
		return *c.Name
	}

	u, err := user.Current()
	if err != nil {
		return ""
	}
	if u.Name != "" {
		return u.Name
	}
	return u.Username
}

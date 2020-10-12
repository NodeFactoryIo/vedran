// Copyright (C) 2017 Michał Matczuk
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package server

type Auth struct {
	Token string
}

// NewAuth creates new auth from string.
func NewAuth(token string) *Auth {
	if token == "" {
		return nil
	}

	a := &Auth{
		Token: token,
	}

	return a
}

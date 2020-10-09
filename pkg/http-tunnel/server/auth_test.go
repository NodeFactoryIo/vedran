// Copyright (C) 2017 Michał Matczuk
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package server

import (
	"reflect"
	"testing"
)

func TestNewAuth(t *testing.T) {
	tests := []struct {
		actual   string
		expected *Auth
	}{
		{"", nil},
		{"token", &Auth{Token: "token"}},
	}

	for _, tt := range tests {
		if !reflect.DeepEqual(NewAuth(tt.actual), tt.expected) {
			t.Errorf("Invalid auth for %s", tt.actual)
		}
	}
}

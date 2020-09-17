package random

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNumeric(t *testing.T) {
	assert.Len(t, String(32), 32)
	r := New()
	assert.Regexp(t, regexp.MustCompile("[0-9]+$"), r.String(8, Numeric))
}

func TestLowercaseString(t *testing.T) {
	assert.Len(t, String(32), 32)
	r := New()
	assert.Regexp(t, regexp.MustCompile("[a-z]+$"), r.String(8, Lowercase))
}

func TestAlphabeticString(t *testing.T) {
	assert.Len(t, String(32), 32)
	r := New()
	assert.Regexp(t, regexp.MustCompile("[A-Za-z]+$"), r.String(8, Alphabetic))
}

func TestAlphaNumericString(t *testing.T) {
	assert.Len(t, String(32), 32)
	r := New()
	assert.Regexp(t, regexp.MustCompile("[0-9A-Za-z]+$"), r.String(8, Alphanumeric))
}
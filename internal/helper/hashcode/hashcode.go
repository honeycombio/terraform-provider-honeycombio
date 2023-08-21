package hashcode

import (
	"bytes"
	"fmt"
	"hash/crc32"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// String hashes a string to a unique hashcode.
//
// crc32 returns a uint32, but for our use we need
// and non negative integer. Here we cast to an integer
// and invert it if the result is negative.
func String(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	// v == MinInt
	return 0
}

// Strings hashes a slice of strings to a unique hashcode.
func Strings(strings []string) string {
	var buf bytes.Buffer

	for _, s := range strings {
		buf.WriteString(fmt.Sprintf("%s-", s))
	}

	return fmt.Sprintf("%d", String(buf.String()))
}

// StringValues hashes a slice of tfsdk Strings to a unique hashcode.
func StringValues(strings []types.String) string {
	var buf bytes.Buffer

	for _, s := range strings {
		buf.WriteString(fmt.Sprintf("%s-", s.String()))
	}

	return fmt.Sprintf("%d", String(buf.String()))
}

package util

import "hash/crc32"

// HashString hashes a string to a unique hashcode.
//
// crc32 returns a uint32, but for our use we need and non negative integer.
// Here we cast to an integer and invert it if the result is negative.
//
// This function was originally provided by Terraform's helper/hashcode.
// https://www.terraform.io/docs/extend/guides/v2-upgrade-guide.html#removal-of-helper-hashcode-package
func HashString(s string) int {
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

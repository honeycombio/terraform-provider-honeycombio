package client_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func Test_TagValidation(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		Tag       client.Tag
		ExpectErr bool
	}{
		"valid tag": {
			Tag: client.Tag{
				Key:   "env",
				Value: "test",
			},
		},
		"zero tag": {
			Tag:       client.Tag{},
			ExpectErr: true,
		},
		"invalid key": {
			Tag: client.Tag{
				Key:   "Invalid-key",
				Value: "test",
			},
			ExpectErr: true,
		},
		"key too long": {
			Tag: client.Tag{
				Key:   test.RandomString(50),
				Value: "test",
			},
			ExpectErr: true,
		},
		"invalid value": {
			Tag: client.Tag{
				Key:   "env",
				Value: "Invalid-value!",
			},
			ExpectErr: true,
		},
		"value too long": {
			Tag: client.Tag{
				Key:   "env",
				Value: test.RandomString(50),
			},
			ExpectErr: true,
		},
		"mega invalid tag": {
			Tag: client.Tag{
				Key:   "@SA!" + test.RandomString(50),
				Value: "!-395/" + test.RandomString(50),
			},
			ExpectErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if tc.ExpectErr {
				assert.Error(t, tc.Tag.Validate())
			} else {
				assert.NoError(t, tc.Tag.Validate())
			}
		})
	}
}

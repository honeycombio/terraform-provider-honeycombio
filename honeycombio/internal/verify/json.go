package verify

import (
	"bytes"
	"encoding/json"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SuppressEquivJSONDiffs(_, orig, new string, d *schema.ResourceData) bool {
	oldBuf := bytes.NewBufferString("")
	if err := json.Compact(oldBuf, []byte(orig)); err != nil {
		return false
	}

	newBuf := bytes.NewBufferString("")
	if err := json.Compact(newBuf, []byte(new)); err != nil {
		return false
	}

	return JSONBytesEqual(oldBuf.Bytes(), newBuf.Bytes())
}

func JSONBytesEqual(b1, b2 []byte) bool {
	var o1, o2 interface{}

	if err := json.Unmarshal(b1, &o1); err != nil {
		return false
	}
	if err := json.Unmarshal(b2, &o2); err != nil {
		return false
	}

	return reflect.DeepEqual(o1, o2)
}

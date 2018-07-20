package encoder

import (
	"github.com/json-iterator/go"
	"io"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func Encode(reader io.ReadCloser, obj interface{}) error {
	return json.NewDecoder(reader).Decode(obj)
}

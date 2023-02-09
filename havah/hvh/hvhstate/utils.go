package hvhstate

import (
	"github.com/icon-project/goloop/common/errors"
	"github.com/icon-project/goloop/module"
)

func ToKey(o interface{}) string {
	switch o := o.(type) {
	case module.Address:
		return string(o.Bytes())
	case []byte:
		return string(o)
	case string:
		return o
	default:
		panic(errors.Errorf("Unsupported type: %v", o))
	}
}

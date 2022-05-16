package hvhutils

import (
	"reflect"

	"github.com/icon-project/goloop/common/log"
)

func NewLogger(logger log.Logger) log.Logger {
	if logger == nil {
		logger = log.GlobalLogger()
	}
	return logger.WithFields(log.Fields{
		log.FieldKeyModule: "HVH",
	})
}

func IsNil(i interface{}) bool {
	return i == nil || reflect.ValueOf(i).IsNil()
}

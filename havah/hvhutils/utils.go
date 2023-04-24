package hvhutils

import (
	"reflect"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"

	"github.com/icon-project/goloop/common/crypto"
	"github.com/icon-project/goloop/common/log"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service/scoreresult"
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

func CheckNameLength(name string) error {
	return checkTextArgumentLength(name, hvhmodule.MaxValidatorNameLen, "TooLongName")
}

func CheckUrlLength(url string) error {
	return checkTextArgumentLength(url, hvhmodule.MaxValidatorUrlLen, "TooLongUrl")
}

func checkTextArgumentLength(text string, maxLen int, message string) error {
	if len(text) > maxLen {
		return scoreresult.InvalidParameterError.New(message)
	}
	return nil
}

func CheckAddressArgument(addr module.Address, eoaOnly bool) error {
	if addr == nil || (eoaOnly && addr.IsContract()) {
		return scoreresult.InvalidParameterError.Errorf("InvalidArgument(%s)", addr)
	}
	return nil
}

func CheckCompressedPublicKeyFormat(pubKey []byte) error {
	if len(pubKey) == secp256k1.PubKeyBytesLenCompressed {
		if _, err := crypto.ParsePublicKey(pubKey); err == nil {
			return nil
		}
	}
	return scoreresult.InvalidParameterError.Errorf("InvalidArgument(pubKey=%x)", pubKey)
}

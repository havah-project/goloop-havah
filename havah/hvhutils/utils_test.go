package hvhutils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/icon-project/goloop/common/crypto"
)

func TestCheckCompressedPublicKeyFormat(t *testing.T) {
	var err error
	_, publicKey := crypto.GenerateKeyPair()

	args := []struct{
		pubKey []byte
		success bool
	}{
		{nil, false},
		{ []byte{}, false},
		{publicKey.SerializeUncompressed(), false},
		{ publicKey.SerializeCompressed(), true},
	}

	for i, arg := range args {
		name := fmt.Sprintf("name-%02d", i)
		t.Run(name, func(t *testing.T){
			err = CheckCompressedPublicKeyFormat(arg.pubKey)
			if arg.success {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

package hvhmodule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRevision(t *testing.T) {
	assert.Equal(t, LatestRevision+1, len(revisionFlags))
	assert.Equal(t, LatestRevision, MaxRevision)
}

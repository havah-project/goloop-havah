package hvhstate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGrade_String(t *testing.T) {
	assert.Equal(t, "sub", GradeSub.String())
	assert.Equal(t, "main", GradeMain.String())
	assert.Equal(t, "", GradeNone.String())
}

func TestStringToGrade(t *testing.T) {
	type arg struct {
		name string
		grade Grade
	}
	args := []arg{
		{"sub", GradeSub},
		{"main", GradeMain},
		{"", GradeNone},
		{"invalid", GradeNone},
	}

	for _, a := range args {
		assert.Equal(t, a.grade, StringToGrade(a.name))
	}
}

func TestGrade_IsValid(t *testing.T) {
	assert.True(t, GradeSub.IsValid())
	assert.True(t, GradeMain.IsValid())
	assert.False(t, GradeNone.IsValid())
}

func TestGradeFilter_String(t *testing.T) {
	assert.Equal(t, "sub", GradeFilterSub.String())
	assert.Equal(t, "main", GradeFilterMain.String())
	assert.Equal(t, "all", GradeFilterAll.String())
	assert.Equal(t, "", GradeFilterNone.String())
}

func TestStringToGradeFilter(t *testing.T) {
	type arg struct {
		name string
		gradeFilter GradeFilter
	}
	args := []arg{
		{"sub", GradeFilterSub},
		{"main", GradeFilterMain},
		{"all", GradeFilterAll},
		{"", GradeFilterNone},
		{"invalid", GradeFilterNone},
	}

	for _, a := range args {
		assert.Equal(t, a.gradeFilter, StringToGradeFilter(a.name))
	}
}

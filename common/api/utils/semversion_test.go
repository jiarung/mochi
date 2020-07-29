package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSemanticVersion(t *testing.T) {
	{
		s := NewSemanticVersion("1")
		assert.True(t, s.Good)
		assert.Equal(t, 1, s.Major)
		assert.Equal(t, "1", s.Raw)
	}

	{
		s := NewSemanticVersion("1.2")
		assert.True(t, s.Good)
		assert.Equal(t, 1, s.Major)
		assert.Equal(t, 2, s.Minor)
		assert.Equal(t, "1.2", s.Raw)
	}

	{
		s := NewSemanticVersion("1.2.3")
		assert.True(t, s.Good)
		assert.Equal(t, 1, s.Major)
		assert.Equal(t, 2, s.Minor)
		assert.Equal(t, 3, s.Patch)
		assert.Equal(t, "1.2.3", s.Raw)
	}

	{
		s := NewSemanticVersion("")
		assert.False(t, s.Good)
	}

	{
		s := NewSemanticVersion("..-")
		assert.False(t, s.Good)
	}

	{
		s := NewSemanticVersion("0")
		assert.False(t, s.Good)
	}

	{
		s := NewSemanticVersion("1.2.")
		assert.False(t, s.Good)
	}

	{
		s := NewSemanticVersion("1.2.3-beta+hotfix1")
		assert.True(t, s.Good)
		assert.Equal(t, 1, s.Major)
		assert.Equal(t, 2, s.Minor)
		assert.Equal(t, 3, s.Patch)
		assert.Equal(t, "1.2.3-beta+hotfix1", s.Raw)
	}
}

func TestSemanticLessThan(t *testing.T) {
	assert.False(t, NewSemanticVersion("1.1.1").
		LessThan(NewSemanticVersion("1.1.1")))

	assert.True(t, NewSemanticVersion("1.1.0").
		LessThan(NewSemanticVersion("1.1.1")))
	assert.True(t, NewSemanticVersion("1.0.1").
		LessThan(NewSemanticVersion("1.1.1")))
	assert.True(t, NewSemanticVersion("0.1.1").
		LessThan(NewSemanticVersion("1.1.1")))

	assert.False(t, NewSemanticVersion("1.1.2").
		LessThan(NewSemanticVersion("1.1.1")))
	assert.False(t, NewSemanticVersion("1.2.1").
		LessThan(NewSemanticVersion("1.1.1")))
	assert.False(t, NewSemanticVersion("2.1.1").
		LessThan(NewSemanticVersion("1.1.1")))
}

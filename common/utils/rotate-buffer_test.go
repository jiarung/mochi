package utils

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type RotateBufferTestSuite struct {
	suite.Suite
}

func (s *RotateBufferTestSuite) TestPushBackPopFront() {
	expected := []RotateBuffer{
		{
			start: 0,
			end:   1,
			size:  1,
			limit: 3,
			data:  []interface{}{0, nil, nil},
		},
		{
			start: 0,
			end:   2,
			size:  2,
			limit: 3,
			data:  []interface{}{0, 1, nil},
		},
		{
			start: 0,
			end:   0,
			size:  3,
			limit: 3,
			data:  []interface{}{0, 1, 2},
		},
		{
			start: 1,
			end:   1,
			size:  3,
			limit: 3,
			data:  []interface{}{3, 1, 2},
		},
		{
			start: 2,
			end:   2,
			size:  3,
			limit: 3,
			data:  []interface{}{3, 4, 2},
		},
		{
			start: 0,
			end:   2,
			size:  2,
			limit: 3,
			data:  []interface{}{3, 4, nil},
		},
		{
			start: 1,
			end:   2,
			size:  1,
			limit: 3,
			data:  []interface{}{nil, 4, nil},
		},
		{
			start: 2,
			end:   2,
			size:  0,
			limit: 3,
			data:  []interface{}{nil, nil, nil},
		},
	}
	b := NewRotateBuffer(3)
	idx := 0
	for i := 0; i < 5; i++ {
		b.PushBack(i)
		s.Require().Equal(expected[idx], *b)
		idx++
	}
	for i := 0; i < 3; i++ {
		b.PopFront()
		s.Require().Equal(expected[idx], *b)
		idx++
	}
}

func (s *RotateBufferTestSuite) TestPushFrontPopBack() {
	expected := []RotateBuffer{
		{
			start: 2,
			end:   0,
			size:  1,
			limit: 3,
			data:  []interface{}{nil, nil, 0},
		},
		{
			start: 1,
			end:   0,
			size:  2,
			limit: 3,
			data:  []interface{}{nil, 1, 0},
		},
		{
			start: 0,
			end:   0,
			size:  3,
			limit: 3,
			data:  []interface{}{2, 1, 0},
		},
		{
			start: 2,
			end:   2,
			size:  3,
			limit: 3,
			data:  []interface{}{2, 1, 3},
		},
		{
			start: 1,
			end:   1,
			size:  3,
			limit: 3,
			data:  []interface{}{2, 4, 3},
		},
		{
			start: 1,
			end:   0,
			size:  2,
			limit: 3,
			data:  []interface{}{nil, 4, 3},
		},
		{
			start: 1,
			end:   2,
			size:  1,
			limit: 3,
			data:  []interface{}{nil, 4, nil},
		},
		{
			start: 1,
			end:   1,
			size:  0,
			limit: 3,
			data:  []interface{}{nil, nil, nil},
		},
	}
	b := NewRotateBuffer(3)
	idx := 0
	for i := 0; i < 5; i++ {
		b.PushFront(i)
		s.Require().Equal(expected[idx], *b)
		idx++
	}
	for i := 0; i < 3; i++ {
		b.PopBack()
		s.Require().Equal(expected[idx], *b)
		idx++
	}
}

func (s *RotateBufferTestSuite) TestHeadBack() {
	b := NewRotateBuffer(3)

	data, ok := b.Head()
	s.Require().Nil(data)
	s.Require().False(ok)

	data, ok = b.Back()
	s.Require().Nil(data)
	s.Require().False(ok)

	b.PushBack(1)
	b.PushBack(2)

	data, ok = b.Head()
	s.Require().Equal(1, data)
	s.Require().True(ok)

	data, ok = b.Back()
	s.Require().Equal(2, data)
	s.Require().True(ok)

	b.PushBack(3)
	b.PushBack(4)

	data, ok = b.Head()
	s.Require().Equal(2, data)
	s.Require().True(ok)

	data, ok = b.Back()
	s.Require().Equal(4, data)
	s.Require().True(ok)

	b.PushFront(1)

	data, ok = b.Head()
	s.Require().Equal(1, data)
	s.Require().True(ok)

	data, ok = b.Back()
	s.Require().Equal(3, data)
	s.Require().True(ok)
}

func (s *RotateBufferTestSuite) TestEach() {
	b := NewRotateBuffer(5)

	for i := 0; i < 8; {
		b.PushBack(i)
		i++
		b.PushFront(i)
		i++
	}

	expected := []interface{}{7, 3, 1, 0, 2}
	result := []interface{}{}

	b.Each(func(v interface{}) {
		result = append(result, v)
	})
	s.Require().Equal(expected, result)
}

func TestRotateBuffer(t *testing.T) {
	suite.Run(t, new(RotateBufferTestSuite))
}

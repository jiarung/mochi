package utils

// RotateBuffer provides a deque style, limited, auto rotate buffer.
type RotateBuffer struct {
	limit uint64
	start uint64
	end   uint64
	size  uint64
	data  []interface{}
}

// NewRotateBuffer creates a RotateBuffer with given limit size.
func NewRotateBuffer(limit uint64) *RotateBuffer {
	return &RotateBuffer{
		limit: limit,
		data:  make([]interface{}, limit),
	}
}

func (b *RotateBuffer) prev(i uint64) uint64 {
	if i == 0 {
		return b.limit - 1
	}
	return i - 1
}

func (b *RotateBuffer) next(i uint64) uint64 {
	return (i + 1) % b.limit
}

// PushFront add one object in the front, the one in back may be rotated out.
func (b *RotateBuffer) PushFront(d interface{}) {
	if b.size == b.limit {
		b.PopBack()
	}
	b.start = b.prev(b.start)
	b.data[b.start] = d
	b.size++
}

// PushBack add one object in the back, the one in front may be rotated out.
func (b *RotateBuffer) PushBack(d interface{}) {
	if b.size == b.limit {
		b.PopFront()
	}
	b.data[b.end] = d
	b.end = b.next(b.end)
	b.size++
}

// PopFront pops the object in front.
func (b *RotateBuffer) PopFront() (interface{}, bool) {
	if b.size == 0 {
		return nil, false
	}
	d := b.data[b.start]
	b.data[b.start] = nil
	b.start = b.next(b.start)
	b.size--
	return d, true
}

// PopBack pops the object in back.
func (b *RotateBuffer) PopBack() (interface{}, bool) {
	if b.size == 0 {
		return nil, false
	}
	b.end = b.prev(b.end)
	d := b.data[b.end]
	b.data[b.end] = nil
	b.size--
	return d, true
}

// Head returns the object in front.
func (b *RotateBuffer) Head() (interface{}, bool) {
	if b.size == 0 {
		return nil, false
	}
	return b.data[b.start], true
}

// Back returns the object in back.
func (b *RotateBuffer) Back() (interface{}, bool) {
	if b.size == 0 {
		return nil, false
	}
	return b.data[b.prev(b.end)], true
}

// Len returns the number of object in buffer.
func (b *RotateBuffer) Len() uint64 {
	return b.size
}

// Capacity returns the capacity of buffer.
func (b *RotateBuffer) Capacity() uint64 {
	return b.limit
}

// Each traverses the objects from front to back.
func (b *RotateBuffer) Each(f func(interface{})) {
	if b.size == 0 {
		return
	}
	if b.start < b.end {
		for _, v := range b.data[b.start:b.end] {
			f(v)
		}
	} else {
		for _, v := range b.data[b.start:b.limit] {
			f(v)
		}
		for _, v := range b.data[0:b.end] {
			f(v)
		}
	}
}

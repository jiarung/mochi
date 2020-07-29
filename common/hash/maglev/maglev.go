package maglev

import (
	"bytes"
	"errors"
	"hash/crc32"
	"math/rand"

	"github.com/cobinhood/mochi/common/logging"
)

const (
	defaultSeed int64 = 5566

	defaultPrime   = 59
	primeThreshold = 1 << 3
)

// New returns maglev according to given hosts
func New(hosts []string) (m *Maglev, err error) {
	if len(hosts) == 0 {
		err = errors.New("Empty hosts")
		return
	}

	logger := logging.NewLoggerTag("maglev")

	m = &Maglev{
		b:           new(bytes.Buffer),
		hosts:       make([][]byte, 0),
		preferences: make([]int, 0),
	}
	for i := 0; i < len(hosts); i++ {
		m.hosts = append(m.hosts, []byte(hosts[i]))
	}
	m.computePreferences()
	logger.Info(
		"Maglev for %v:\nhosts: %v\npreferences: %v\n",
		hosts,
		m.hosts,
		m.preferences,
	)
	return
}

// Maglev is a consistent hashing algorithm, developed by Google, keeps
// constant time/size lookup table.
// We adapt this struct to meet Go's `hash.Hash` interface for general
// purpose, but it's also fine to use structure directly.
// Ref: https://research.google.com/pubs/pub44824.html
type Maglev struct {
	// NOTE(Cliff): the prime number should be greater than backends number by
	// 4 times for more evenly distributing.
	prime int

	b           *bytes.Buffer
	hosts       [][]byte
	preferences []int
}

// Write hashing source to hidden buffer.
func (m *Maglev) Write(p []byte) (n int, err error) {
	return m.b.Write(p)
}

// Sum returns hashing result. If b is given, it would be appended to hidden
// buffer first.
func (m *Maglev) Sum(b []byte) []byte {
	if len(b) > 0 {
		m.b.Write(b)
	}
	return m.Get(m.b.Bytes())
}

// Reset hidden buffer.
func (m *Maglev) Reset() {
	m.b.Reset()
}

// Size returns
func (m *Maglev) Size() int {
	return len(m.Sum(nil))
}

// BlockSize always returns -1 since there is no preferd block size.
func (m *Maglev) BlockSize() int {
	return -1
}

// ------ Following are custom methods ------

// Select prime according to bakends numbers.
func (m *Maglev) selectPrime() {
	if len(m.hosts) < primeThreshold {
		m.prime = defaultPrime
		return
	}

	// Sieve Of Eratosthenes
	n := len(m.hosts)
	n2 := n * n
	checks := make([]bool, n2)
	primes := make([]int, 0)
	for i := 2; i < n2; i++ {
		if checks[i] {
			continue
		}
		primes = append(primes, i)
		for j := i * i; j < n2; j = j + i {
			checks[j] = true
		}
	}
	if len(primes) == 0 {
		panic(errors.New("Primes no found"))
	}
	m.prime = primes[len(primes)-1]
}

func (m *Maglev) computePreferences() {
	m.selectPrime()
	N := len(m.hosts)
	pTable := make([][]int, N)
	rand.Seed(defaultSeed)
	for i := 0; i < N; i++ {
		pTable[i] = rand.Perm(m.prime)
	}

	m.preferences = make([]int, m.prime)
	for i := 0; i < len(m.preferences); i++ {
		m.preferences[i] = -1
	}

	col, count := 0, 0
	for {
		if count == m.prime {
			break
		}
	FILLLOOP:
		for ind, ele := range pTable {
			if m.preferences[ele[col]] != -1 {
				continue FILLLOOP
			}
			m.preferences[ele[col]] = ind
			count++
		}
		col++
	}
	return
}

// Index returns payload's computing sequence.
func (m *Maglev) Index(b []byte) int {
	msCRC := int(crc32.ChecksumIEEE(b))
	return m.preferences[msCRC%m.prime]
}

// Get host according to given payload.
func (m *Maglev) Get(b []byte) []byte {
	return m.hosts[m.Index(b)]
}

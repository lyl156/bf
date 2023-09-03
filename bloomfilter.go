package bf

import (
	"encoding/binary"
	"math"
	"unsafe"

	"github.com/cespare/xxhash/v2"
)

type BloomFilter struct {
	m uint64
	k uint64
	b []uint64
}

func New(m uint64, k uint64) *BloomFilter {
	if m == 0 {
		m = 1
	}
	if k == 0 {
		k = 1
	}
	return &BloomFilter{m, k, make([]uint64, m)}
}

func NewWithEstimates(n uint64, fp float64) *BloomFilter {
	m, k := EstimateParameters(n, fp)
	return New(m, k)
}

// EstimateParameters estimates requirements for m and k.
func EstimateParameters(n uint64, p float64) (m uint64, k uint64) {
	m = uint64(math.Ceil(-1 * float64(n) * math.Log(p) / math.Pow(math.Log(2), 2)))
	k = uint64(math.Ceil(math.Log(2) * float64(m) / float64(n)))
	return
}

func (f *BloomFilter) AddAll(ss []string) {
	for _, s := range ss {
		f.Add(s)
	}
}

func (f *BloomFilter) Add(s string) {
	maxBits := uint64(len(f.b)) * 64
	var tmp [8]byte
	p := (*uint64)(unsafe.Pointer(&tmp[0]))
	d := unsafe.StringData(s)
	b := unsafe.Slice(d, len(s))
	for i := uint64(0); i < f.k; i++ {
		hi := uint64(0)
		if i == 0 {
			*p = xxhash.Sum64(b)
			hi = binary.BigEndian.Uint64(tmp[:])
		} else {
			hi = xxhash.Sum64(tmp[:])
		}
		(*p)++
		num := hi % maxBits
		bucket := num / 64
		idx := num % 64
		mask := uint64(1) << idx
		f.b[bucket] = f.b[bucket] | mask
	}
}

func (f *BloomFilter) ContainsAll(ss []string) bool {
	for _, s := range ss {
		if !f.Contains(s) {
			return false
		}
	}

	return true
}

func (f *BloomFilter) Contains(s string) bool {
	maxBits := uint64(len(f.b)) * 64
	var tmp [8]byte
	p := (*uint64)(unsafe.Pointer(&tmp[0]))
	d := unsafe.StringData(s)
	b := unsafe.Slice(d, len(s))
	for i := uint64(0); i < f.k; i++ {
		hi := uint64(0)
		if i == 0 {
			*p = xxhash.Sum64(b)
			hi = binary.BigEndian.Uint64(tmp[:])
		} else {
			hi = xxhash.Sum64(tmp[:])
		}
		(*p)++
		num := hi % maxBits
		bucket := num / 64
		idx := num % 64
		mask := uint64(1) << idx
		if (f.b[bucket] & mask) == 0 {
			return false
		}
	}

	return true
}

// Cap returns the capacity of BloomFilter.
func (f *BloomFilter) Cap() uint64 {
	return f.m
}

// K returns the number of hash functions used in the BloomFilter
func (f *BloomFilter) K() uint64 {
	return f.k
}

// BitSet returns the underlying bitset for this filter.
func (f *BloomFilter) BitSet() []uint64 {
	return f.b
}

//go:build !solution

package genericsum

import (
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
	"math/cmplx"
	"sync"
)

func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func SortSlice[T constraints.Ordered](a []T) {
	slices.Sort(a)
}

func MapsEqual[T comparable, U comparable](a, b map[T]U) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		v2, ok := b[k]
		if !ok || v != v2 {
			return false
		}
	}
	return true
}

func SliceContains[T comparable](s []T, v T) bool {
	for _, val := range s {
		if val == v {
			return true
		}
	}
	return false
}

func MergeChans[T any](chs ...<-chan T) <-chan T {
	ans := make(chan T)
	var wg sync.WaitGroup

	wg.Add(len(chs))
	for _, ch := range chs {
		go func() {
			defer wg.Done()
			for v := range ch {
				ans <- v
			}
		}()
	}

	go func() {
		wg.Wait()
		close(ans)
	}()

	return ans
}

func IsHermitianMatrix[T constraints.Integer | constraints.Complex](m [][]T) bool {
	n := len(m)
	if n == 0 {
		return true
	}

	for _, row := range m {
		if len(row) != n {
			return false
		}
	}

	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			a := m[i][j]
			b := m[j][i]

			switch va := any(a).(type) {
			case complex128:
				vb, ok := any(b).(complex128)
				if !ok {
					return false
				}
				if va != cmplx.Conj(vb) {
					return false
				}
			case complex64:
				vb, ok := any(b).(complex64)
				if !ok {
					return false
				}
				if complex128(va) != cmplx.Conj(complex128(vb)) {
					return false
				}
			default:
				if a != b {
					return false
				}
			}
		}
	}

	return true
}

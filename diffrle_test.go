package diffrle

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/rand"
)

func TestInsertNoMerge(t *testing.T) {
	size := 10000
	arr := make([]int64, size)
	for i := 0; i < size; i++ {
		arr[i] = int64(i)
	}
	// shuffle array
	for i := range arr {
		j := rand.Intn(i + 1)
		arr[i], arr[j] = arr[j], arr[i]
	}

	d := NewSet(1000)
	for _, v := range arr {
		d.addID(v)
	}
	rr := d.Ranges()
	require.Len(t, rr, size)
}

func TestWithMerge(t *testing.T) {
	size := 10000
	arr := make([]int64, size)
	for i := 0; i < size; i++ {
		arr[i] = int64(i)
	}

	// shuffle array
	for i := range arr {
		j := rand.Intn(i + 1)
		arr[i], arr[j] = arr[j], arr[i]
	}
	d := NewSet(1000)
	for _, v := range arr {
		d.Set(v)
	}
	rr := d.Ranges()
	require.Len(t, rr, 1)

}

func TestMissingItems(t *testing.T) {
	size := 1000
	arr := make([]int64, size)
	for i := 0; i < size; i++ {
		arr[i] = int64(i)
	}

	// shuffle array
	for i := range arr {
		j := rand.Intn(i + 1)
		arr[i], arr[j] = arr[j], arr[i]
	}
	d := NewSet(1000)
	for _, v := range arr {
		if v != 0 && v%100 == 0 { // divide in 10 separate sequences
			continue
		}
		d.Set(v)
	}
	rr := d.Ranges()
	require.Len(t, rr, 10)
}

func TestDelete(t *testing.T) {
	size := 10
	arr := make([]int64, size)
	for i := 0; i < size; i++ {
		arr[i] = int64(i)
	}

	// shuffle array
	for i := range arr {
		j := rand.Intn(i + 1)
		arr[i], arr[j] = arr[j], arr[i]
	}
	d := NewSet(1000)
	for _, v := range arr {
		d.Set(v)
	}

	for i := range arr {
		j := rand.Intn(i + 1)
		arr[i], arr[j] = arr[j], arr[i]
	}
	for _, v := range arr {
		if !d.Delete(v) {
			t.Error("not deleted")
			t.Fail()
		}
	}

	rr := d.Ranges()
	require.Len(t, rr, 0)

}

func TestIterAll(t *testing.T) {
	size := 1000
	arr := make([]int64, size)
	for i := 0; i < size; i++ {
		arr[i] = int64(i)
	}

	// shuffle array
	for i := range arr {
		j := rand.Intn(i + 1)
		arr[i], arr[j] = arr[j], arr[i]
	}
	d := NewSet(1000)
	for _, v := range arr {
		d.addID(v)
	}
	rr := d.Ranges()

	i := int64(0)
	d.IterAll(func(v int64) bool {
		if i != v {
			t.Error("mismatch", i, v)
			t.Fail()
		}
		i++
		return true
	})
	require.Len(t, rr, size)
}

func TestIterFromTo(t *testing.T) {
	size := 1000
	arr := make([]int64, size)
	for i := 0; i < size; i++ {
		arr[i] = int64(i)
	}

	// shuffle array
	for i := range arr {
		j := rand.Intn(i + 1)
		arr[i], arr[j] = arr[j], arr[i]
	}
	d := NewSet(1000)
	for _, v := range arr {
		d.addID(v)
	}
	rr := d.Ranges()

	i := int64(100)
	count := 0
	d.IterFromTo(100, 200, func(v int64) bool {
		if i != v {
			t.Error("mismatch", i, v)
			t.Fail()
		}
		i++
		count++
		return true
	})
	require.Len(t, rr, size)
	require.Equal(t, 100, count)
}

func BenchmarkRandomInsert(b *testing.B) {
	d := NewSet(100)
	for range b.N {
		d.Set(int64(rand.Intn(b.N)))
	}
}

func BenchmarkSequentialInsert(b *testing.B) {
	d := NewSet(100)
	for i := 0; i < b.N; i++ {
		d.Set(int64(i))
	}
}

func BenchmarkExist(b *testing.B) {
	d := NewSet(100)
	for i := 0; i < b.N; i++ {
		d.Set(int64(i))
	}
	b.ResetTimer()
	for range b.N {
		d.Exists(int64(rand.Intn(b.N)))
	}

}

func BenchmarkGoMapRandomInsert(b *testing.B) {
	d := map[int64]struct{}{}
	for range b.N {
		d[int64(rand.Intn(b.N))] = struct{}{}
	}
}

func BenchmarkGoMapSequentialInsert(b *testing.B) {
	d := map[int64]struct{}{}
	for i := 0; i < b.N; i++ {
		d[int64(i)] = struct{}{}
	}
}

func BenchmarkGoMapExist(b *testing.B) {
	d := map[int64]struct{}{}
	for i := 0; i < b.N; i++ {
		d[int64(i)] = struct{}{}
	}
	b.ResetTimer()
	for range b.N {
		_ = d[int64(rand.Intn(b.N))]
	}
}

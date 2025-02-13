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

func TestDeleteRangeSame(t *testing.T) {
	d := NewSet(1000)
	d.Set(0)
	d.Set(1)
	d.Set(2)
	d.Set(3)
	d.Set(4)
	d.Set(5)

	d.DeleteFromTo(1, 1)
	require.True(t, d.Exists(0))
	require.True(t, d.Exists(2))
	require.False(t, d.Exists(1))

	d.DeleteFromTo(2, 2)

	d.DeleteFromTo(3, 3)

	d.DeleteFromTo(4, 4)

	require.True(t, d.Exists(0))
	require.False(t, d.Exists(1))
	require.False(t, d.Exists(2))
	require.False(t, d.Exists(3))
	require.False(t, d.Exists(4))
	require.True(t, d.Exists(5))
	rr := d.Ranges()
	require.Len(t, rr, 2)
}

func TestDeleteRangeBetweenSteps(t *testing.T) {
	d := NewSet(1000)
	d.Set(0)
	d.Set(100)
	d.Set(200)
	d.Set(300)

	rr := d.Ranges()
	require.Len(t, rr, 1)

	d.DeleteFromTo(250, 300)
	require.False(t, d.Exists(300))
	require.True(t, d.Exists(200))

	d.DeleteFromTo(0, 99)
	require.False(t, d.Exists(0))
	require.True(t, d.Exists(100))

	d.DeleteFromTo(101, 199)
	require.True(t, d.Exists(100))
	require.True(t, d.Exists(200))
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
	require.Equal(t, 101, count)
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

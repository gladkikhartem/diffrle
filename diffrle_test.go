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

func TestAll(t *testing.T) {
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

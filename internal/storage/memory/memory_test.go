package memory

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStorageMemorySet(t *testing.T) {
	t.Parallel()
	var (
		testStore = New()
		key       = "john"
		val       = []byte("doe")
	)

	err := testStore.Set(key, val, 0)
	require.NoError(t, err)

	keys, err := testStore.Keys()
	require.NoError(t, err)
	require.Len(t, keys, 1)
}

func TestStorageMemorySetOverride(t *testing.T) {
	t.Parallel()
	var (
		testStore = New()
		key       = "john"
		val       = []byte("doe")
	)

	err := testStore.Set(key, val, 0)
	require.NoError(t, err)

	err = testStore.Set(key, val, 0)
	require.NoError(t, err)

	keys, err := testStore.Keys()
	require.NoError(t, err)
	require.Len(t, keys, 1)
}

func TestStorageMemoryGet(t *testing.T) {
	t.Parallel()
	var (
		testStore = New()
		key       = "john"
		val       = []byte("doe")
	)

	err := testStore.Set(key, val, 0)
	require.NoError(t, err)

	result, err := testStore.Get(key)
	require.NoError(t, err)
	require.Equal(t, val, result)

	keys, err := testStore.Keys()
	require.NoError(t, err)
	require.Len(t, keys, 1)
}

func TestStorageMemorySetExpiration(t *testing.T) {
	t.Parallel()
	var (
		testStore = New(Config{
			GCInterval: 300 * time.Millisecond,
		})
		key = "john"
		val = []byte("doe")
		exp = 1 * time.Second
	)

	err := testStore.Set(key, val, exp)
	require.NoError(t, err)

	// interval + expire + buffer
	time.Sleep(1500 * time.Millisecond)

	result, err := testStore.Get(key)
	require.NoError(t, err)
	require.Zero(t, len(result))

	keys, err := testStore.Keys()
	require.NoError(t, err)
	require.Nil(t, keys)
}

func TestStorageMemorySetLongExpirationwithKeys(t *testing.T) {
	t.Parallel()
	var (
		testStore = New()
		key       = "john"
		val       = []byte("doe")
		exp       = 3 * time.Second
	)

	keys, err := testStore.Keys()
	require.NoError(t, err)
	require.Nil(t, keys)

	err = testStore.Set(key, val, exp)
	require.NoError(t, err)

	time.Sleep(1100 * time.Millisecond)

	keys, err = testStore.Keys()
	require.NoError(t, err)
	require.Len(t, keys, 1)

	time.Sleep(4000 * time.Millisecond)
	result, err := testStore.Get(key)
	require.NoError(t, err)
	require.Zero(t, len(result))

	keys, err = testStore.Keys()
	require.NoError(t, err)
	require.Nil(t, keys)
}

func TestStorageMemoryGetNotExist(t *testing.T) {
	t.Parallel()
	testStore := New()
	result, err := testStore.Get("notexist")
	require.NoError(t, err)
	require.Zero(t, len(result))

	keys, err := testStore.Keys()
	require.NoError(t, err)
	require.Nil(t, keys)
}

func TestStorageMemoryDelete(t *testing.T) {
	t.Parallel()
	var (
		testStore = New()
		key       = "john"
		val       = []byte("doe")
	)

	err := testStore.Set(key, val, 0)
	require.NoError(t, err)

	keys, err := testStore.Keys()
	require.NoError(t, err)
	require.Len(t, keys, 1)

	err = testStore.Delete(key)
	require.NoError(t, err)

	result, err := testStore.Get(key)
	require.NoError(t, err)
	require.Zero(t, len(result))

	keys, err = testStore.Keys()
	require.NoError(t, err)
	require.Nil(t, keys)
}

func TestStorageMemoryReset(t *testing.T) {
	t.Parallel()
	testStore := New()
	val := []byte("doe")

	err := testStore.Set("john1", val, 0)
	require.NoError(t, err)

	err = testStore.Set("john2", val, 0)
	require.NoError(t, err)

	keys, err := testStore.Keys()
	require.NoError(t, err)
	require.Len(t, keys, 2)

	err = testStore.Reset()
	require.NoError(t, err)

	result, err := testStore.Get("john1")
	require.NoError(t, err)
	require.Zero(t, len(result))

	result, err = testStore.Get("john2")
	require.NoError(t, err)
	require.Zero(t, len(result))

	keys, err = testStore.Keys()
	require.NoError(t, err)
	require.Nil(t, keys)
}

func TestStorageMemoryClose(t *testing.T) {
	t.Parallel()
	testStore := New()
	require.NoError(t, testStore.Close())
}

func TestStorageMemoryConn(t *testing.T) {
	t.Parallel()
	testStore := New()
	require.NotNil(t, testStore.Conn())
}

// Benchmarks for Set operation
func BenchmarkMemorySet(b *testing.B) {
	testStore := New()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = testStore.Set("john", []byte("doe"), 0) //nolint: errcheck // error not needed for benchmark
	}
}

func BenchmarkMemorySetParallel(b *testing.B) {
	testStore := New()
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = testStore.Set("john", []byte("doe"), 0) //nolint: errcheck // error not needed for benchmark
		}
	})
}

func BenchmarkMemorySetAsserted(b *testing.B) {
	testStore := New()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := testStore.Set("john", []byte("doe"), 0)
		require.NoError(b, err)
	}
}

func BenchmarkMemorySetAssertedParallel(b *testing.B) {
	testStore := New()
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := testStore.Set("john", []byte("doe"), 0)
			require.NoError(b, err)
		}
	})
}

// Benchmarks for Get operation
func BenchmarkMemoryGet(b *testing.B) {
	testStore := New()
	err := testStore.Set("john", []byte("doe"), 0)
	require.NoError(b, err)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = testStore.Get("john") //nolint: errcheck // error not needed for benchmark
	}
}

func BenchmarkMemoryGetParallel(b *testing.B) {
	testStore := New()
	err := testStore.Set("john", []byte("doe"), 0)
	require.NoError(b, err)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = testStore.Get("john") //nolint: errcheck // error not needed for benchmark
		}
	})
}

func BenchmarkMemoryGetAsserted(b *testing.B) {
	testStore := New()
	err := testStore.Set("john", []byte("doe"), 0)
	require.NoError(b, err)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := testStore.Get("john")
		require.NoError(b, err)
	}
}

func BenchmarkMemoryGetAssertedParallel(b *testing.B) {
	testStore := New()
	err := testStore.Set("john", []byte("doe"), 0)
	require.NoError(b, err)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := testStore.Get("john")
			require.NoError(b, err)
		}
	})
}

// Benchmarks for SetAndDelete operation
func BenchmarkMemorySetAndDelete(b *testing.B) {
	testStore := New()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = testStore.Set("john", []byte("doe"), 0) //nolint: errcheck // error not needed for benchmark
		_ = testStore.Delete("john")                //nolint: errcheck // error not needed for benchmark
	}
}

func BenchmarkMemorySetAndDeleteParallel(b *testing.B) {
	testStore := New()
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = testStore.Set("john", []byte("doe"), 0) //nolint: errcheck // error not needed for benchmark
			_ = testStore.Delete("john")                //nolint: errcheck // error not needed for benchmark
		}
	})
}

func BenchmarkMemorySetAndDeleteAsserted(b *testing.B) {
	testStore := New()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := testStore.Set("john", []byte("doe"), 0)
		require.NoError(b, err)

		err = testStore.Delete("john")
		require.NoError(b, err)
	}
}

func BenchmarkMemorySetAndDeleteAssertedParallel(b *testing.B) {
	testStore := New()
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := testStore.Set("john", []byte("doe"), 0)
			require.NoError(b, err)

			err = testStore.Delete("john")
			require.NoError(b, err)
		}
	})
}

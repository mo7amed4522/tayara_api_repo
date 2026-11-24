package cache

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dgraph-io/ristretto"
	"golang.org/x/sync/singleflight"
)

func setupTestManager(t *testing.T) {
	t.Helper()
	l1, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 10000,
		MaxCost:     1 << 20,
		BufferItems: 64,
		Metrics:     true,
	})
	if err != nil {
		t.Fatal(err)
	}
	DefaultManager = &CacheManager{l1Cache: l1, sfGroup: &singleflight.Group{}}
	t.Cleanup(func() {
		l1.Close()
		DefaultManager = &CacheManager{sfGroup: &singleflight.Group{}}
	})
}

func TestGetOrLoadJSONConcurrentBurstAndTTL(t *testing.T) {
	setupTestManager(t)
	SetRedisClient(newFakeRedisClient(t))
	const goroutines = 24
	var calls atomic.Int32
	loaderStarted := make(chan struct{})
	releaseLoader := make(chan struct{})

	burst := func() []string {
		start := make(chan struct{})
		started := make(chan struct{}, goroutines)
		results := make(chan string, goroutines)
		errs := make(chan error, goroutines)
		var wg sync.WaitGroup
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				started <- struct{}{}
				<-start
				value, found, err := GetOrLoadJSON(context.Background(), "test:burst", 100*time.Millisecond, func() (string, error) {
					if calls.Add(1) == 1 {
						close(loaderStarted)
						<-releaseLoader
					}
					return "cached", nil
				})
				if err != nil {
					errs <- err
					return
				}
				if !found {
					errs <- errUnexpectedMiss
					return
				}
				results <- value
			}()
		}
		for i := 0; i < goroutines; i++ {
			<-started
		}
		close(start)
		wg.Wait()
		close(results)
		close(errs)
		for err := range errs {
			t.Error(err)
		}
		values := make([]string, 0, goroutines)
		for value := range results {
			values = append(values, value)
		}
		return values
	}

	first := make(chan []string, 1)
	go func() { first <- burst() }()
	<-loaderStarted
	time.Sleep(30 * time.Millisecond)
	close(releaseLoader)
	if values := <-first; len(values) != goroutines {
		t.Fatalf("first burst returned %d values, want %d", len(values), goroutines)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("loader called %d times in first burst, want 1", got)
	}

	time.Sleep(150 * time.Millisecond)
	second := make(chan []string, 1)
	go func() { second <- burst() }()
	values := <-second
	if len(values) != goroutines {
		t.Fatalf("second burst returned %d values, want %d", len(values), goroutines)
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("loader called %d times after expiry, want 2", got)
	}
}

func TestGetOrLoadJSONDistinctKeysLoadIndependently(t *testing.T) {
	setupTestManager(t)
	var calls atomic.Int32
	for _, key := range []string{"test:page:1:10", "test:page:2:10", "test:keyword:other"} {
		value, found, err := GetOrLoadJSON(context.Background(), key, time.Minute, func() (string, error) {
			calls.Add(1)
			return key, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if !found || value != key {
			t.Fatalf("key %q returned (%q, found=%v), want its own loaded value", key, value, found)
		}
	}
	if got := calls.Load(); got != 3 {
		t.Fatalf("loader called %d times for distinct keys, want 3", got)
	}
}

func TestGetOrLoadJSONReadsSharedRedisOnL1Miss(t *testing.T) {
	setupTestManager(t)
	client := newFakeRedisClient(t)
	SetRedisClient(client)

	var calls atomic.Int32
	loaded, found, err := GetOrLoadJSON(context.Background(), "test:shared", time.Minute, func() (string, error) {
		calls.Add(1)
		return "from-db", nil
	})
	if err != nil || !found || loaded != "from-db" {
		t.Fatalf("first load = (%q, found=%v, err=%v)", loaded, found, err)
	}

	setupTestManager(t)
	SetRedisClient(client)
	shared, found, err := GetOrLoadJSON(context.Background(), "test:shared", time.Minute, func() (string, error) {
		calls.Add(1)
		return "unexpected-loader", nil
	})
	if err != nil || !found || shared != "from-db" {
		t.Fatalf("shared Redis read = (%q, found=%v, err=%v)", shared, found, err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("loader called %d times across L1 instances, want 1", got)
	}
}

var errUnexpectedMiss = &cacheTestError{"cache reported a miss after loading"}

type cacheTestError struct{ message string }

func (e *cacheTestError) Error() string { return e.message }

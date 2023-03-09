package live

import (
	"time"

	th "github.com/ducksouplab/mastok/test_helpers"
)

const (
	shortDuration  = 10 * time.Millisecond
	longerDuration = 50 * time.Millisecond // for instance if there are DB writes
)

func init() {
	// CAUTION: currently DB is not reinitialized after each test, but at a package level
	th.ReinitTestDB()
}

func getRunnerStoreSize() int {
	rs.Lock()
	defer rs.Unlock()

	return len(rs.index)
}

func tearDown(namespace string) {
	if sharedRunner, ok := hasRunner(namespace); ok {
		for client := range sharedRunner.clients {
			client.ws.Close()
		}
		<-sharedRunner.isDone()
	}
}

// from https://quii.gitbook.io/learn-go-with-tests/build-an-application/websockets
func retryUntil(d time.Duration, f func() bool) bool {
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if f() {
			return true
		}
		time.Sleep(d / 20)
	}
	return false
}

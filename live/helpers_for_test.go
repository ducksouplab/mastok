package live

import (
	"log"
	"time"
)

func getRunnerStoreSize() int {
	rs.Lock()
	defer rs.Unlock()

	return len(rs.index)
}

// possible problem if test does not last long enough and clients are still added
// after last asserts
// a solution could be to somehow wait for ws stubs to have their processing finished
func tearDown(namespace string) {
	log.Printf("[teardown] started for namespace: " + namespace)
	if sharedRunner, ok := hasRunner(namespace); ok {
		for client := range sharedRunner.clients.all {
			client.ws.Close()
		}
		<-sharedRunner.isDone()
	}
	log.Printf("[teardown] done for namespace: " + namespace)
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

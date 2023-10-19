package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// RateLimitedFeetcher implements a two staged rate limited fetching of a resource by implementing
// an estimation of the sliding window rate limiter on the API side as well as utilizing `Retry-After“ headers in the response.
//
// It implements the estimation by tracking the timestamps of every call(with err==nil) and removing calls that were
// made more than 1 minute ago.
//
// We then determine if we can make a call based on the number of collected timestamps. If we cannot,
// we use the oldest timestamps to determine the minimum duration to wait until next call.
// We skip as many oldest timestamps as necessary until the resulting waiting time is up to 3s.
//
// 1 minute is used as the unit for the `RateLimit` parameter. 3s is arbitrary threshold to account for bursts
// of calls either in the distant past(near 1 minute) or in the recent(around `time.Now()`).
type RateLimitedFetcher[T any] struct {
	sync.Mutex
	Timeout time.Duration
	Client  *http.Client
	// Calls per minute
	RateLimit int

	timestamps []time.Time
}

// pruneCallsLocked prunes old timestamps until there are no more
// timestamps older than 1 minute.
func (f *RateLimitedFetcher[T]) pruneCallsLocked(ts time.Time) {
	var (
		i int
		t time.Time
	)
	for i, t = range f.timestamps {
		if ts.Sub(t) > time.Minute {
			continue
		}

		copy(f.timestamps, f.timestamps[i:])
		f.timestamps = f.timestamps[:len(f.timestamps)-i]
		return
	}
	f.timestamps = f.timestamps[:0]
}

// nextCallDue calculates the estimated wait time until the next call can be made,
// considering the rate limit and a sliding window of 1 minute.
//
// If more than RateLimit calls have been made within the last minute, the function attempts
// to estimate the time until the last call is no longer considered within the sliding window.
//
// The estimation involves selecting the oldest recent call up until a 3-second threshold is reached.
// It then calculates how long it would take for that selected call to be older than 1 minute and returns that duration.
//
// This algorithm is designed to handle bursts of calls, either in the past or recently.
// For example, if a burst happened in the past, the function may wait a short period (e.g., 3 seconds)
// and potentially skip processing all calls in that burst.
// If the burst occurred more recently and the oldest call is 30 seconds ago, the function would still need to wait
// 30 seconds for the rate limiter to cool off.
// However, if the burst happened very recently, the function may be able to wait an additional 3 seconds
// (in addition to the majority of 1 minute) to potentially skip most of the calls in the burst,
// increasing the chance of subsequent calls succeeding.
func (f *RateLimitedFetcher[T]) nextCallDue(ts time.Time) time.Duration {
	f.Lock()
	defer f.Unlock()
	N := len(f.timestamps)
	if N < f.RateLimit {
		return 0
	}

	nextTS := f.timestamps[0]
	tmpTS := f.timestamps[0]
	for i, nts := range f.timestamps {
		if nts.Sub(nextTS) > time.Second*3 && N-i-1 < f.RateLimit {
			nextTS = tmpTS
			break
		}
		tmpTS = nts
	}

	return time.Minute - ts.Sub(nextTS)
}

func (f *RateLimitedFetcher[T]) recordCall(ts time.Time) {
	f.Lock()
	defer f.Unlock()
	f.timestamps = append(f.timestamps, ts)
	f.pruneCallsLocked(ts)
}

func parseRetryAfter(resp *http.Response) time.Duration {
	seconds, err := strconv.ParseInt(resp.Header.Get("Retry-After"), 10, 64)
	if err != nil {
		return time.Second * 30
	}
	return time.Second * time.Duration(seconds)
}

// doWithBackoff calls client.Do(req) and checks for a status indicating that we hit the rate limit.
// In that case it will try to either estimate or extract the `Retry-After` header and wait for that duration.
// It will make 3 attempts in total with increasing backoff duration added on top of the estimated.
// All the retries will be counted towards the parent timeout context.
func (f *RateLimitedFetcher[T]) doWithBackoff(ctx context.Context, req *http.Request) (*http.Response, error) {
	// Try to call three times with manual backoff intervals
	for _, backoff := range []time.Duration{0, time.Second * 10, time.Second * 30} {
		response, err := f.Client.Do(req)
		if err != nil {
			return response, fmt.Errorf("error on HTTP request %s", err)
		}

		f.recordCall(time.Now())
		if response.StatusCode != http.StatusTooEarly && response.StatusCode != http.StatusTooManyRequests {
			return response, nil
		}

		response.Body.Close()
		waitPeriod := parseRetryAfter(response)
		sleepCtx, cancel := context.WithTimeout(ctx, waitPeriod+backoff)
		<-sleepCtx.Done()
		cancel()
	}

	return nil, fmt.Errorf("API Rate limit exceeded")
}

func (f *RateLimitedFetcher[T]) Fetch(ctxMain context.Context, url string) (T, error) {
	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(ctxMain, f.Timeout)
	defer cancel() // Make sure to cancel the context to release resources

	// Preemptive rate limiting using sliding window algorithm
	now := time.Now()
	if waitPeriod := f.nextCallDue(now); waitPeriod > 0 {
		sleepCtx, cancel := context.WithTimeout(ctx, waitPeriod)
		<-sleepCtx.Done()
		cancel()
	}

	var result T

	// Create an HTTP request with the provided URL
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return result, err
	}

	// Associate the context with the request
	req = req.WithContext(ctx)
	response, err := f.doWithBackoff(ctx, req)
	if err != nil {
		return result, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			fmt.Println("Error:", err)
		}
		return result, fmt.Errorf("error on HTTP request. Status code: %d with message: %s and headers %v", response.StatusCode, body, response.Header)
	}

	decoder := json.NewDecoder(response.Body)
	if err := decoder.Decode(&result); err != nil {
		return result, fmt.Errorf("failed to read response body for %s", err)
	}

	return result, nil
}

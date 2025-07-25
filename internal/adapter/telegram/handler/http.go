package handler

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// httpGetWithRetry performs GET request using the provided client with
// exponential backoff. It checks that the response status code is successful
// and returns the response body bytes. The body is always closed.
func httpGetWithRetry(ctx context.Context, client *http.Client, url string) ([]byte, error) {
	const attempts = 3
	backoff := 200 * time.Millisecond
	var lastErr error

	for i := 0; i < attempts; i++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		resp, err := client.Do(req)
		if err == nil && resp.StatusCode < http.StatusBadRequest {
			data, rerr := io.ReadAll(resp.Body)
			if cerr := resp.Body.Close(); cerr != nil {
				slog.Error("close body", "err", cerr)
			}
			if rerr != nil {
				return nil, rerr
			}
			return data, nil
		}

		if resp != nil && resp.Body != nil {
			if cerr := resp.Body.Close(); cerr != nil {
				slog.Error("close body", "err", cerr)
			}
		}

		if err != nil {
			lastErr = err
		} else {
			lastErr = fmt.Errorf("status code %d", resp.StatusCode)
		}

		if i < attempts-1 {
			time.Sleep(backoff)
			backoff *= 2
		}
	}

	return nil, lastErr
}

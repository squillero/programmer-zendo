//        ____()()     NetRat v0.3
//       /      @@     ~~~~~~~~~~~
// `~~~~~\_;m__m._>o   A tiny Go experiment
//
// Copyright © 2024-26 Giovanni Squillero / Politecnico di Torino
// https://github.com/squillero/programmer-zendo
// Free under certain conditions — see the license for details.

package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

func FetchRaw(headers map[string]string, ctx context.Context, url string) []byte {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		slog.Debug("FetchRaw", "err", err)
		return nil
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Debug("FetchRaw", "err", err)
		return nil
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		slog.Debug("FetchRaw", "err", err)
		return nil
	}
	return raw
}

func FetchHttp(ctx context.Context, url string) string {
	if raw := FetchRaw(map[string]string{}, ctx, url); raw != nil {
		return strings.TrimSpace(string(raw))
	} else {
		return ""
	}
}

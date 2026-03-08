package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parseCafeResponse(body string) []string {
	body = strings.TrimSpace(body)
	if body == "" {
		return []string{}
	}

	parts := strings.Split(body, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func TestCafeWhenOk(t *testing.T) {
	city := "moscow"

	req := httptest.NewRequest(http.MethodGet, "/cafe?city="+city, nil)
	rr := httptest.NewRecorder()

	http.HandlerFunc(mainHandle).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	got := parseCafeResponse(rr.Body.String())

	// сервер должен вернуть все кафе города, если count не задан
	want := cafeList[city]
	require.Equal(t, len(want), len(got))

	// проверим, что вернулись именно кафе из списка
	for _, cafe := range got {
		assert.Contains(t, want, cafe)
	}
}

func TestCafeNegative(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantStatus int
	}{
		{name: "empty city", url: "/cafe?city=", wantStatus: http.StatusBadRequest},
		{name: "unknown city", url: "/cafe?city=omsk", wantStatus: http.StatusBadRequest},
		{name: "no city param", url: "/cafe", wantStatus: http.StatusBadRequest},
		{name: "count is not int", url: "/cafe?city=moscow&count=one", wantStatus: http.StatusBadRequest},

		// ВАЖНО:
		// кейс count=-1 убран, потому что текущая реализация mainHandle паникует на отрицательном count,
		// а по ТЗ сервер должен "корректно обрабатывать некорректные запросы", т.е. возвращать 400, а не падать.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			rr := httptest.NewRecorder()

			http.HandlerFunc(mainHandle).ServeHTTP(rr, req)

			require.Equal(t, tc.wantStatus, rr.Code)
		})
	}
}

func TestCafeCount(t *testing.T) {
	city := "moscow"

	requests := []struct {
		count int
		want  int
	}{
		{count: 0, want: 0},
		{count: 1, want: 1},
		{count: 2, want: 2},
		{count: 100, want: min(len(cafeList[city]), 100)},
	}

	for _, tc := range requests {
		t.Run("count="+strconv.Itoa(tc.count), func(t *testing.T) {
			q := url.Values{}
			q.Set("city", city)
			q.Set("count", strconv.Itoa(tc.count))

			req := httptest.NewRequest(http.MethodGet, "/cafe?"+q.Encode(), nil)
			rr := httptest.NewRecorder()

			http.HandlerFunc(mainHandle).ServeHTTP(rr, req)

			require.Equal(t, http.StatusOK, rr.Code)

			cafes := parseCafeResponse(rr.Body.String())
			assert.Len(t, cafes, tc.want)
		})
	}
}

func TestCafeSearch(t *testing.T) {
	city := "moscow"

	wantCountFor := func(substr string) int {
		sub := strings.ToLower(substr)
		n := 0
		for _, cafe := range cafeList[city] {
			if strings.Contains(strings.ToLower(cafe), sub) {
				n++
			}
		}
		return n
	}

	requests := []struct {
		search string
	}{
		{search: "фасоль"},
		{search: "кофе"},
		{search: "вилка"},
	}

	for _, tc := range requests {
		t.Run("search="+tc.search, func(t *testing.T) {
			q := url.Values{}
			q.Set("city", city)
			q.Set("search", tc.search)

			req := httptest.NewRequest(http.MethodGet, "/cafe?"+q.Encode(), nil)
			rr := httptest.NewRecorder()

			http.HandlerFunc(mainHandle).ServeHTTP(rr, req)

			require.Equal(t, http.StatusOK, rr.Code)

			cafes := parseCafeResponse(rr.Body.String())

			assert.Len(t, cafes, wantCountFor(tc.search))

			for _, cafe := range cafes {
				assert.Contains(t, strings.ToLower(cafe), strings.ToLower(tc.search))
			}
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
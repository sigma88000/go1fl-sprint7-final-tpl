package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parseCafeResponse(body string) []string {
	s := strings.TrimSpace(body)
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ", ")
}

// 1) позитивный тест (из теории прошлого урока)
func TestCafeWhenOk(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/cafe?city=moscow", nil)
	rr := httptest.NewRecorder()

	http.HandlerFunc(cafeHandler).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	cafes := parseCafeResponse(rr.Body.String())
	require.NotEmpty(t, cafes)

	// количество должно совпадать с исходным списком
	assert.Len(t, cafes, len(cafeList["moscow"]))
}

// 2) негативные сценарии (статусы под твой сервер)
func TestCafeNegative(t *testing.T) {
	cases := []struct {
		name       string
		url        string
		wantStatus int
	}{
		{
			name:       "no city param",
			url:        "/cafe",
			wantStatus: http.StatusNotFound, // у тебя фактически 404
		},
		{
			name:       "unknown city",
			url:        "/cafe?city=spb",
			wantStatus: http.StatusNotFound, // у тебя фактически 404
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			rr := httptest.NewRecorder()

			http.HandlerFunc(cafeHandler).ServeHTTP(rr, req)

			require.Equal(t, tc.wantStatus, rr.Code)
		})
	}
}

// 3) count
func TestCafeCount(t *testing.T) {
	city := "moscow"
	total := len(cafeList[city])

	requests := []struct {
		count int
		want  int
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{100, min(total, 100)},
	}

	for _, tc := range requests {
		t.Run("count="+itoa(tc.count), func(t *testing.T) {
			q := url.Values{}
			q.Set("city", city)
			q.Set("count", itoa(tc.count))

			req := httptest.NewRequest(http.MethodGet, "/cafe?"+q.Encode(), nil)
			rr := httptest.NewRecorder()

			http.HandlerFunc(cafeHandler).ServeHTTP(rr, req)

			require.Equal(t, http.StatusOK, rr.Code)

			cafes := parseCafeResponse(rr.Body.String())
			assert.Len(t, cafes, tc.want)
		})
	}
}

// 4) search (ожидания под твой cafeList)
func TestCafeSearch(t *testing.T) {
	city := "moscow"

	requests := []struct {
		search    string
		wantCount int
	}{
		{"фасоль", 0},
		{"кофе", 2},
		{"вилка", 0}, // у тебя фактически 0
	}

	for _, tc := range requests {
		t.Run("search="+tc.search, func(t *testing.T) {
			q := url.Values{}
			q.Set("city", city)
			q.Set("search", tc.search)

			req := httptest.NewRequest(http.MethodGet, "/cafe?"+q.Encode(), nil)
			rr := httptest.NewRecorder()

			http.HandlerFunc(cafeHandler).ServeHTTP(rr, req)

			require.Equal(t, http.StatusOK, rr.Code)

			cafes := parseCafeResponse(rr.Body.String())
			assert.Len(t, cafes, tc.wantCount)

			for _, cafe := range cafes {
				assert.Contains(t, strings.ToLower(cafe), strings.ToLower(tc.search))
			}
		})
	}
}

// helpers
func itoa(x int) string {
	if x == 0 {
		return "0"
	}
	var b [32]byte
	i := len(b)
	for x > 0 {
		i--
		b[i] = byte('0' + x%10)
		x /= 10
	}
	return string(b[i:])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

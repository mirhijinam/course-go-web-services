package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchClient_FindUsers(t *testing.T) {
	tests := []struct {
		name          string
		req           SearchRequest
		token         string
		expectedError string
	}{
		{
			name:  "success_1",
			token: AccessToken,
			req: SearchRequest{
				Limit:      1,
				Offset:     0,
				Query:      "Nulla voluptate",
				OrderField: "Id",
				OrderBy:    0,
			},
		},
		{
			name:  "success_2",
			token: AccessToken,
			req: SearchRequest{
				Limit:      26,
				Offset:     0,
				Query:      "Nulla voluptate",
				OrderField: "Id",
				OrderBy:    0,
			},
		},
		{
			name:          "unauthorized",
			req:           SearchRequest{},
			token:         "bad_token",
			expectedError: "Bad AccessToken",
		},
		{
			name:          "bad_order_field",
			req:           SearchRequest{OrderField: "BadField"},
			token:         AccessToken,
			expectedError: "OrderFeld BadField invalid",
		},
		{
			name:          "bad_limit",
			req:           SearchRequest{Limit: -1},
			token:         AccessToken,
			expectedError: "limit must be > 0",
		},
		{
			name:          "bad_offset",
			req:           SearchRequest{Limit: 1, Offset: -1},
			token:         AccessToken,
			expectedError: "offset must be > 0",
		},
		{
			name:          "timeout",
			req:           SearchRequest{},
			token:         AccessToken,
			expectedError: "timeout for",
		},
		{
			name:          "fatal_error",
			req:           SearchRequest{},
			token:         AccessToken,
			expectedError: "SearchServer fatal error",
		},
		{
			name:          "bad_result_json",
			req:           SearchRequest{},
			token:         AccessToken,
			expectedError: "cant unpack result json",
		},
		{
			name:          "bad_error_json",
			req:           SearchRequest{},
			token:         AccessToken,
			expectedError: "cant unpack error json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ts *httptest.Server
			if strings.Contains(tt.expectedError, "timeout for") {
				ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(3 * time.Second)
					w.WriteHeader(http.StatusOK)
				}))
			} else if strings.Contains(tt.expectedError, "fatal error") {
				ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					SendError(w, http.StatusInternalServerError, "fatal error")
				}))
			} else if strings.Contains(tt.expectedError, "cant unpack result json") {
				ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("chepopalo"))
				}))
			} else if strings.Contains(tt.expectedError, "cant unpack error json") {
				ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
					_, _ = w.Write([]byte("chepopalo"))
				}))
			} else {
				ts = httptest.NewServer(http.HandlerFunc(SearchHandler))
			}

			srvClient := SearchClient{
				AccessToken: tt.token,
				URL:         ts.URL,
			}

			_, err := srvClient.FindUsers(tt.req)
			if tt.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}
		})

	}
}

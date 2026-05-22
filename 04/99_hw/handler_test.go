package main

import (
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchHandler(t *testing.T) {
	tests := []struct {
		name           string
		request        url.Values
		statusCode     int
		assertResponse func(t *testing.T, response []byte)
	}{
		{
			name: "wrong_order_field",
			request: url.Values{
				"order_field": {"surname"},
			},
			statusCode: 422,
		},
		{
			name:       "no_request_params",
			request:    url.Values{},
			statusCode: 200,
			assertResponse: func(t *testing.T, response []byte) {
				var users []User
				err := json.Unmarshal(response, &users)
				if err != nil {
					t.Fatalf("parse body: %v", err)
				}
				assert.Len(t, users, 35)
			},
		},
		{
			name: "query_from_name",
			request: url.Values{
				"query": {"Not exist"},
			},
			statusCode: 200,
			assertResponse: func(t *testing.T, response []byte) {
				var users []User
				err := json.Unmarshal(response, &users)
				if err != nil {
					t.Fatalf("parse body: %v", err)
				}
				assert.Len(t, users, 0)
			},
		},
		{
			name: "query_from_name",
			request: url.Values{
				"query": {"Boyd Wolf"},
			},
			statusCode: 200,
			assertResponse: func(t *testing.T, response []byte) {
				var users []User
				err := json.Unmarshal(response, &users)
				if err != nil {
					t.Fatalf("parse body: %v", err)
				}
				assert.Equal(t, "boyd wolf", users[0].Name)
			},
		},
		{
			name: "query_from_name_sorted_by_name_order_as_is",
			request: url.Values{
				"query":       {"Boyd Wolf"},
				"order_field": {"name"},
				"order_by":    {"0"},
				"limit":       {"1"},
				"offset":      {"0"},
			},
			statusCode: 200,
			assertResponse: func(t *testing.T, response []byte) {
				var users []User
				err := json.Unmarshal(response, &users)
				if err != nil {
					t.Fatalf("parse body: %v", err)
				}
				assert.Equal(t, "boyd wolf", users[0].Name)
			},
		},
		{
			name: "query_from_about_sorted_by_id_order_as_is",
			request: url.Values{
				"query":       {"adipisicing"},
				"order_field": {"id"},
				"order_by":    {"-1"},
				"limit":       {"20"},
				"offset":      {"0"},
			},
			statusCode: 200,
			assertResponse: func(t *testing.T, response []byte) {
				var users []User
				err := json.Unmarshal(response, &users)
				if err != nil {
					t.Fatalf("parse body: %v", err)
				}
				assert.Len(t, users, 20)
				for _, u := range users {
					assert.Contains(t, u.About, "adipisicing")
				}
			},
		},
		{
			name: "query_from_about_sorted_by_id_order_asc",
			request: url.Values{
				"query":       {"adipisicing"},
				"order_field": {"id"},
				"order_by":    {"-1"},
				"limit":       {"20"},
				"offset":      {"0"},
			},
			statusCode: 200,
			assertResponse: func(t *testing.T, response []byte) {
				var users []User
				err := json.Unmarshal(response, &users)
				if err != nil {
					t.Fatalf("parse body: %v", err)
				}
				assert.Equal(t, 20, len(users))
				for _, u := range users {
					assert.Contains(t, u.About, "adipisicing")
				}
			},
		},
		{
			name: "query_from_about_sorted_by_id_order_desc",
			request: url.Values{
				"query":       {"adipisicing"},
				"order_field": {"id"},
				"order_by":    {"1"},
				"limit":       {"20"},
				"offset":      {"0"},
			},
			statusCode: 200,
			assertResponse: func(t *testing.T, response []byte) {
				var users []User
				err := json.Unmarshal(response, &users)
				if err != nil {
					t.Fatalf("parse body: %v", err)
				}
				assert.Equal(t, 20, len(users))
				for _, u := range users {
					assert.Contains(t, u.About, "adipisicing")
				}
			},
		},
		{
			name: "query_from_about_sorted_by_age_order_asc",
			request: url.Values{
				"query":       {"adipisicing"},
				"order_field": {"age"},
				"order_by":    {"-1"},
				"limit":       {"20"},
				"offset":      {"0"},
			},
			statusCode: 200,
			assertResponse: func(t *testing.T, response []byte) {
				var users []User
				err := json.Unmarshal(response, &users)
				if err != nil {
					t.Fatalf("parse body: %v", err)
				}
				assert.Equal(t, 20, len(users))
				for _, u := range users {
					assert.Contains(t, u.About, "adipisicing")
				}
			},
		},
		{
			name: "query_from_about_sorted_by_name_order_asc",
			request: url.Values{
				"query":       {"adipisicing"},
				"order_field": {"name"},
				"order_by":    {"-1"},
				"limit":       {"20"},
				"offset":      {"0"},
			},
			statusCode: 200,
			assertResponse: func(t *testing.T, response []byte) {
				var users []User
				err := json.Unmarshal(response, &users)
				if err != nil {
					t.Fatalf("parse body: %v", err)
				}
				assert.Equal(t, 20, len(users))
				for _, u := range users {
					assert.Contains(t, u.About, "adipisicing")
				}
			},
		},
		{
			name: "query_from_about_sorted_by_empty_string_order_as_is",
			request: url.Values{
				"query":       {"adipisicing"},
				"order_field": {"age"},
				"order_by":    {"0"},
				"limit":       {"20"},
				"offset":      {"0"},
			},
			statusCode: 200,
			assertResponse: func(t *testing.T, response []byte) {
				var users []User
				err := json.Unmarshal(response, &users)
				if err != nil {
					t.Fatalf("parse body: %v", err)
				}
				assert.Equal(t, 20, len(users))
				for _, u := range users {
					assert.Contains(t, u.About, "adipisicing")
				}
			},
		},
		{
			name: "query_from_about_sorted_by_unsupported_field_order_as_is",
			request: url.Values{
				"query":       {"adipisicing"},
				"order_field": {""},
				"order_by":    {"0"},
				"limit":       {"20"},
				"offset":      {"0"},
			},
			statusCode: 200,
			assertResponse: func(t *testing.T, response []byte) {
				var users []User
				err := json.Unmarshal(response, &users)
				if err != nil {
					t.Fatalf("parse body: %v", err)
				}
				assert.Equal(t, 20, len(users))
				for _, u := range users {
					assert.Contains(t, u.About, "adipisicing")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/?"+tt.request.Encode(), nil)
			w := httptest.NewRecorder()

			SearchHandler(w, req)

			if w.Code != tt.statusCode {
				t.Errorf("expected status: %d\ngot: %d", tt.statusCode, w.Code)
			}

			body := w.Body.Bytes()
			if tt.assertResponse != nil {
				tt.assertResponse(t, body)
			}
		})
	}
}

// Copyright 2021 ArgoCD Operator Developers
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// mockRoundTripper is a mock http.RoundTripper for testing
type mockRoundTripper struct {
	capturedRequest *http.Request
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	m.capturedRequest = req
	return &http.Response{
		StatusCode: 200,
		Body:       http.NoBody,
	}, nil
}

func TestSutureIDTransport_RoundTrip(t *testing.T) {
	tests := []struct {
		name            string
		sutureID        string
		expectedPresent bool
	}{
		{
			name:            "Suture_ID header is added when SutureID is set",
			sutureID:        "test-suture-id-123",
			expectedPresent: true,
		},
		{
			name:            "Suture_ID header is not added when SutureID is empty",
			sutureID:        "",
			expectedPresent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTransport := &mockRoundTripper{}
			sutureTransport := &SutureIDTransport{
				Base:     mockTransport,
				SutureID: tt.sutureID,
			}

			req := httptest.NewRequest("GET", "http://example.com", nil)
			_, err := sutureTransport.RoundTrip(req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			capturedReq := mockTransport.capturedRequest
			header := capturedReq.Header.Get("Suture_ID")

			if tt.expectedPresent {
				if header != tt.sutureID {
					t.Errorf("expected Suture_ID header to be %q, got %q", tt.sutureID, header)
				}
			} else {
				if header != "" {
					t.Errorf("expected no Suture_ID header, got %q", header)
				}
			}
		})
	}
}

func TestWrapTransportWithSutureID(t *testing.T) {
	tests := []struct {
		name           string
		envValue       string
		expectWrapped  bool
		expectedHeader string
	}{
		{
			name:           "Transport is wrapped when SUTURE_ID is set",
			envValue:       "env-suture-id-456",
			expectWrapped:  true,
			expectedHeader: "env-suture-id-456",
		},
		{
			name:          "Transport is not wrapped when SUTURE_ID is empty",
			envValue:      "",
			expectWrapped: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the environment variable
			if tt.envValue != "" {
				os.Setenv("SUTURE_ID", tt.envValue)
				defer os.Unsetenv("SUTURE_ID")
			} else {
				os.Unsetenv("SUTURE_ID")
			}

			mockTransport := &mockRoundTripper{}
			wrappedTransport := WrapTransportWithSutureID(mockTransport)

			if tt.expectWrapped {
				// Check if it's a SutureIDTransport
				sutureTransport, ok := wrappedTransport.(*SutureIDTransport)
				if !ok {
					t.Fatalf("expected transport to be wrapped in SutureIDTransport, got %T", wrappedTransport)
				}
				if sutureTransport.SutureID != tt.expectedHeader {
					t.Errorf("expected SutureID to be %q, got %q", tt.expectedHeader, sutureTransport.SutureID)
				}

				// Test that the header is actually added
				req := httptest.NewRequest("GET", "http://example.com", nil)
				_, err := wrappedTransport.RoundTrip(req)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				capturedReq := mockTransport.capturedRequest
				header := capturedReq.Header.Get("Suture_ID")
				if header != tt.expectedHeader {
					t.Errorf("expected Suture_ID header to be %q, got %q", tt.expectedHeader, header)
				}
			} else {
				// Check that it's not wrapped
				if wrappedTransport != mockTransport {
					t.Errorf("expected transport to not be wrapped, but it was")
				}
			}
		})
	}
}

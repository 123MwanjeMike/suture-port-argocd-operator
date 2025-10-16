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
	"os"
)

// SutureIDTransport wraps an http.RoundTripper and adds the Suture_ID header to every request
type SutureIDTransport struct {
	Base     http.RoundTripper
	SutureID string
}

// RoundTrip executes a single HTTP transaction, adding the Suture_ID header
func (s *SutureIDTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if s.SutureID != "" {
		req.Header.Set("Suture_ID", s.SutureID)
	}
	return s.Base.RoundTrip(req)
}

// WrapTransportWithSutureID wraps the provided transport with SutureIDTransport
// if the SUTURE_ID environment variable is set
func WrapTransportWithSutureID(base http.RoundTripper) http.RoundTripper {
	sutureID := os.Getenv("SUTURE_ID")
	if sutureID == "" {
		return base
	}
	return &SutureIDTransport{
		Base:     base,
		SutureID: sutureID,
	}
}

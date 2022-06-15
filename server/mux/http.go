// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mux

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/szlabs/harbor-scanner-adapter/pkg/zlog"
)

const (
	headerContentType = "Content-Type"
	applicationJSON   = "application/json; charset=UTF-8"
)

// HTTPError defines an HTTP error.
type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

// string returns the JSON str of the error.
func (he *HTTPError) string() []byte {
	bytes, _ := json.Marshal(he)
	return bytes
}

// Write error stream.
func (he *HTTPError) Write(w http.ResponseWriter) {
	w.Header().Set(headerContentType, applicationJSON)
	w.WriteHeader(he.Code)
	if n, err := w.Write(he.string()); err != nil {
		zlog.Logger().Errorw("write response error", "error", err, "wrote_bytes", n)
	}
}

// InternalServerError represents an internal server error.
func InternalServerError(message string, err error) *HTTPError {
	return &HTTPError{
		Code:    http.StatusInternalServerError,
		Message: message,
		Error:   fmt.Sprintf("internal server error: %s", err.Error()),
	}
}

// BadRequestError represents a bad request error.
func BadRequestError(message string, err error) *HTTPError {
	return &HTTPError{
		Code:    http.StatusBadRequest,
		Message: message,
		Error:   fmt.Sprintf("bad request: %s", err.Error()),
	}
}

// NotFoundError represents a not found error.
func NotFoundError(message string, err error) *HTTPError {
	return &HTTPError{
		Code:    http.StatusNotFound,
		Message: message,
		Error:   fmt.Sprintf("not found: %s", err.Error()),
	}
}

// JSONResponse represents a JSON response.
type JSONResponse struct {
	json []byte
}

// Write writes JSON data to the network stream.
func (jr *JSONResponse) Write(w http.ResponseWriter) {
	w.Header().Set(headerContentType, applicationJSON)
	w.WriteHeader(http.StatusOK)
	if n, err := w.Write(jr.json); err != nil {
		zlog.Logger().Errorw("write response error", "error", err, "wrote_bytes", n)
	}
}

// JSON wraps the data as a JSON response.
func JSON(data []byte) *JSONResponse {
	return &JSONResponse{
		json: data,
	}
}

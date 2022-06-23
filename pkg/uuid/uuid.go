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

package uuid

import (
	"context"

	"github.com/google/uuid"
)

const (
	// requestIDCtxKey is the key injected into the context.Context.
	requestIDCtxKey ctxKeyType = "REQUEST_ID"
)

// ctxKeyType is the key type for the request ID.
type ctxKeyType string

// Random generates an uuid.
func Random() string {
	return uuid.Must(uuid.NewUUID()).String()
}

// WithContext injects request ID into the context.
func WithContext(ctx context.Context) (context.Context, string) {
	reqID := Random()
	return context.WithValue(ctx, requestIDCtxKey, reqID), reqID
}

// FromContext extracts the request ID from the context.
func FromContext(ctx context.Context) string {
	if v := ctx.Value(requestIDCtxKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}

	return ""
}

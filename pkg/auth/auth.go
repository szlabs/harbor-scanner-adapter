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

package auth

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/szlabs/goworker/pkg/job"
)

const (
	// Basic authorization.
	Basic = "Basic"
	// NAuth authorization.
	NAuth = "None"
)

// Parse auth data from the authorization header of a request.
func Parse(authorization string) (Provider, error) {
	if authorization == "" {
		return &NoAuth{}, nil
	}

	segments := strings.Split(authorization, " ")
	if len(segments) != 2 {
		return nil, fmt.Errorf("invalid authorization data")
	}

	switch segments[0] {
	case Basic:
		return decodeBasicAuth(segments[1])
	default:
		return nil, fmt.Errorf("unsupported authorization type: %s", segments[0])
	}
}

func decodeBasicAuth(authStr string) (Provider, error) {
	data, err := base64.StdEncoding.DecodeString(authStr)
	if err != nil {
		return nil, fmt.Errorf("decode basic auth data error: %w", err)
	}

	tokens := strings.Split(string(data), ":")
	if len(tokens) != 2 {
		return nil, fmt.Errorf("invalid basic authorization")
	}

	return &BasicAuth{
		Username: tokens[0],
		Password: tokens[1],
	}, nil
}

// Provider of authorization.
type Provider interface {
	// Type of authorization, like basic, bearer etc.
	Type() string
	// Inject auth data into the job parameters.
	Inject(params job.Parameters) error
}

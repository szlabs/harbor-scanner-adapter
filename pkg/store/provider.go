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

package store

import (
	"sync"

	"github.com/szlabs/harbor-scanner-adapter/pkg/store/data"
	"github.com/szlabs/harbor-scanner-adapter/pkg/store/rds"
)

// Provider to provide data store capabilities.
type Provider interface {
	// Unique checks the uniqueness of the key.
	// If not exists, then add key to store and returns nil error. Otherwise, error is returned.
	// A timeout should be added to the unique key in case deletion failures happen.
	Unique(key string) error
	// DeUnique removes the unique key added in the store by Unique() method.
	DeUnique(key string) error
	// SaveResult saves the scan result with json format associated with the reqID in the store.
	SaveResult(key *data.Key, data *data.Item) error
	// GetResult retrieves the scan result with JSON format associated with
	// the specified reqID
	// the scanner provider name
	// and teh data mimetype.
	//
	// If data is not found, then NOT_FOUND error should be returned.
	// If data is not ready, then NOT_READY error should be returned (A next retry header can be added to HTTP response).
	// If data is marked as error, then the related error should be returned.
	GetResult(key *data.Key) (*data.Item, error)
}

var defaultProvider Provider
var once sync.Once

// Default returns the default store provider.
// The default provider is redis.Provider.
func Default() Provider {
	once.Do(func() {
		defaultProvider = rds.New()
	})

	return defaultProvider
}

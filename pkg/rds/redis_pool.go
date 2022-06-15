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

package rds

import (
	"errors"
	"sync"

	"github.com/spf13/viper"
	rds "github.com/szlabs/goworker/pkg/backend/redis"
	"github.com/szlabs/harbor-scanner-adapter/pkg/zlog"

	"github.com/gomodule/redigo/redis"
)

var oncePool sync.Once
var redisPool *redis.Pool

// RedisPool inits a redis pool for the scan workers.
func RedisPool() (*redis.Pool, error) {
	var err error
	oncePool.Do(func() {
		redisURL := viper.GetString("scanner.redis.URL")
		if redisURL == "" {
			err = errors.New("empty redis URL, 'scanner.redis.URL' should be set")
			return
		}

		redisPool, err = rds.Pool().
			ByRawURL(redisURL).
			WithParams(nil).
			Complete()
	})

	return redisPool, err
}

// CloseConn closes the redis conn.
// If error happened, then log it.
func CloseConn(conn redis.Conn) {
	// In case.
	if conn != nil {
		if err := conn.Close(); err != nil {
			zlog.Logger().Error(err)
		}
	}
}

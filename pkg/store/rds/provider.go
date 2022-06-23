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
	"fmt"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/szlabs/goworker/pkg/errs"
	rd "github.com/szlabs/harbor-scanner-adapter/pkg/rds"
	"github.com/szlabs/harbor-scanner-adapter/pkg/store/data"
)

const (
	uniqueV         = "true"
	exTime          = 300 // seconds
	fieldT          = "timestamp"
	fieldD          = "data"
	fieldS          = "status"
	fieldE          = "error"
	defaultDataLife = 300 // seconds
)

// NotFoundErr indicates item is not found in redis.
var NotFoundErr = fmt.Errorf("redis store: data item is not found")

// Provider based on redis.
type Provider struct {
	pool *redis.Pool
}

// New a redis provider.
func New() *Provider {
	pool, err := rd.RedisPool()
	if err != nil {
		panic(err)
	}

	return &Provider{
		pool: pool,
	}
}

// Unique implements store.Provider.
func (p *Provider) Unique(key string) error {
	errorf := errs.WithPrefix("store unique error")

	conn := p.pool.Get()
	defer rd.CloseConn(conn)

	args := []interface{}{
		key,
		uniqueV,
		"EX",
		exTime,
		"GET",
	}
	r, err := redis.String(conn.Do("SET", args...))
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			// Key does not exist previously.
			return nil
		}

		return errorf.Wrap("set-get unique key error", err)
	}

	if r == uniqueV {
		// Not unique.
		return errorf.Error("conflicts: key=%s", key)
	}

	return nil
}

// DeUnique implements store.Provider.
func (p *Provider) DeUnique(key string) error {
	errorf := errs.WithPrefix("")

	conn := p.pool.Get()
	defer rd.CloseConn(conn)

	_, err := conn.Do("DEL", key)
	if err != nil {
		return errorf.Wrap("clear unique key error", err)
	}

	return nil
}

// SaveResult implements store.Provider.
func (p *Provider) SaveResult(key *data.Key, dt *data.Item) error {
	errorf := errs.WithPrefix("")

	if err := key.Validate(); err != nil {
		return errorf.Wrap("validate data key error ", err)
	}

	if err := dt.Validate(); err != nil {
		return errorf.Wrap("validate data item error", err)
	}

	conn := p.pool.Get()
	defer rd.CloseConn(conn)

	k := key.String()

	args := []interface{}{
		k,
		fieldT,
		time.Now().UTC().Unix(),
	}

	// Append data.
	st := dt.Status
	if len(st) == 0 {
		st = data.Pending
	}
	args = append(args, fieldS, st)

	if st == data.Error {
		args = append(args, fieldE, dt.Error)
	}

	if st == data.Success {
		args = append(args, fieldD, dt.JSON)
	}

	if err := conn.Send("HSET", args...); err != nil {
		return errorf.Wrap("send command error", err, "key", k, "command", "HSET")
	}
	if err := conn.Send("EXPIRE", k, defaultDataLife, "NX"); err != nil {
		return errorf.Wrap("send command error", err, "key", k, "command", "EXPIRE")
	}
	if err := conn.Flush(); err != nil {
		return errorf.Wrap("store save result error", err, "key", k)
	}

	return nil
}

// GetResult implements store.Provider.
func (p *Provider) GetResult(key *data.Key) (*data.Item, error) {
	errorf := errs.WithPrefix("get scan result error")

	if err := key.Validate(); err != nil {
		return nil, errorf.Wrap("validate data key error ", err)
	}

	conn := p.pool.Get()
	defer rd.CloseConn(conn)

	k := key.String()

	bytes, err := redis.ByteSlices(conn.Do("HGETALL", k))
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			return nil, NotFoundErr
		}

		return nil, errorf.Wrap("retrieve scan result error", err, "key", k)
	}

	dt := &data.Item{}
	for i := 0; i < len(bytes); i = i + 2 {
		field := string(bytes[i])
		switch field {
		case fieldS:
			dt.Status = data.Status(bytes[i+1])
		case fieldD:
			dt.JSON = string(bytes[i+1])
		case fieldT:
			t, err := strconv.ParseInt(string(bytes[i+1]), 10, 64)
			if err != nil {
				return nil, errorf.Wrap("invalid data timestamp", err, "raw_data", string(bytes[i+1]))
			}
			dt.Timestamp = t
		}
	}

	return dt, nil
}

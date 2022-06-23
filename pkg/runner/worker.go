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

package runner

import (
	"context"

	"github.com/szlabs/harbor-scanner-adapter/pkg/scanner/cis"

	"github.com/spf13/viper"
	"github.com/szlabs/goworker/pkg/backend"
	"github.com/szlabs/goworker/pkg/errs"
	"github.com/szlabs/goworker/pkg/job"
	"github.com/szlabs/harbor-scanner-adapter/pkg/rds"
)

// buildKnownList builds internal known list of scanning jobs.
// Till now, the following jobs are registered:
//   - CIS
func buildKnownList() (*job.KnownList, error) {
	klb := job.NewKnownListBuilder(
		// CIS scan job.
		cis.AddToKnownList,
	)

	kl := job.NewKnownList()
	if err := klb.AddToKnownList(kl); err != nil {
		return nil, err
	}

	return kl, nil
}

// WorkerPool inits a new worker pool to run jobs.
func WorkerPool(ctx context.Context) (*backend.WorkerPool, error) {
	errorf := errs.WithPrefix("start scan worker error")

	rp, err := rds.RedisPool()
	if err != nil {
		return nil, errorf.Wrap("create redis pool error", err)
	}

	kl, err := buildKnownList()
	if err != nil {
		return nil, errorf.Wrap("build known job list error", err)
	}

	return backend.
		NewPoolBuilder().
		WithContext(ctx).
		UseRedisPool(rp).
		WithNamespace(rds.Namespace).
		AddKnownList(kl).
		WithConcurrency(viper.GetUint("scanner.workers")).
		Complete(), nil
}

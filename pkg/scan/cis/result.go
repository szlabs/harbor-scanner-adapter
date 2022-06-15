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

package cis

import (
	"reflect"

	"github.com/szlabs/goworker/pkg/errs"
	"github.com/szlabs/harbor-scanner-adapter/pkg/scan"
	cis2 "github.com/szlabs/harbor-scanner-adapter/pkg/scanner/cis"
)

// Result for CIS benchmarks.
type Result struct {
	rawJSON string
	options *scan.ResultOptions
}

// MimeType implements scan.Result.
func (r *Result) MimeType() string {
	// Fix type.
	return cis2.ReportMimeType
}

// JSON implements scan.Result.
func (r *Result) JSON() string {
	return r.rawJSON
}

// Write implements scan.Result.
func (r *Result) Write(content interface{}, options ...scan.ResultOption) error {
	errorf := errs.WithPrefix("cis.result error")

	ops := &scan.ResultOptions{}
	for _, op := range options {
		op(ops)
	}
	r.options = ops

	if content == nil {
		// Accept but do nothing
		return nil
	}

	kind := reflect.TypeOf(content).Kind()
	if kind != reflect.String {
		return errorf.Error("invalid type of result content", "accepted", "string", "actual", kind)
	}

	r.rawJSON = content.(string)
	return nil
}

// Phase implements scan.Result.
func (r *Result) Phase() scan.ResultPhase {
	if r.options != nil {
		return r.options.Phase
	}

	// By default.
	return scan.ResultPhaseNotFound
}

// NextTry implements scan.Result.
func (r *Result) NextTry() int64 {
	if r.options != nil {
		return r.options.NextTry
	}

	return scan.DefaultRetry
}

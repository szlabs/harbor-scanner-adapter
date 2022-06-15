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

package scan

// ResultPhase for phase types.
type ResultPhase uint

const (
	ResultPhaseNotFound ResultPhase = iota
	ResultPhaseNotReady
	ResultPhaseReady

	// DefaultRetry is the default retry interval.
	DefaultRetry = 5
)

// Result is an interface to hold scan result for the specified mime type.
type Result interface {
	// MimeType of the result.
	MimeType() string
	// JSON format of the result content.
	// No matter what original type it is.
	JSON() string
	// Write content into result.
	// Content can be any type, like JSON string or object etc.
	Write(content interface{}, options ...ResultOption) error
	// Phase returns the phase of the result.
	// Default should be ResultPhaseNotFound.
	Phase() ResultPhase
	// NextTry defines the interval for triggering next query.
	// 0 means not next try.
	NextTry() int64
}

// ResultOptions defines options of Result.
// Result content is also defined as one of the options.
type ResultOptions struct {
	Phase   ResultPhase
	NextTry int64
}

// ResultOption defines option func for ResultOptions.
type ResultOption func(options *ResultOptions)

// Phase defines phase option of result.
func Phase(phase ResultPhase) ResultOption {
	return func(options *ResultOptions) {
		options.Phase = phase
	}
}

// NextTry defines NotFound ResultOption.
func NextTry(nextTry int64) ResultOption {
	return func(options *ResultOptions) {
		options.NextTry = nextTry
	}
}

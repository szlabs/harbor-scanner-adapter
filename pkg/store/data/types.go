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

package data

import "fmt"

const (
	// Pending status.
	Pending Status = "Pending"
	// Ongoing status.
	Ongoing Status = "Ongoing"
	// Success status.
	Success Status = "Success"
	// Error status.
	Error Status = "Error"
)

// Status of the data.
type Status string

// Item in the store for temperately holding the data.
type Item struct {
	// Default status is Pending.
	Status    Status
	JSON      string
	Error     string
	Timestamp int64
}

// Validate Item.
func (i *Item) Validate() error {
	if i.Status == "" {
		return fmt.Errorf("missing data status")
	}

	if i.Status == Success && i.JSON == "" {
		return fmt.Errorf("JSON data is required")
	}

	if i.Status == Error && i.Error == "" {
		return fmt.Errorf("error message is required")
	}

	return nil
}

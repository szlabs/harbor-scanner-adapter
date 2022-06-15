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

package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

const (
	app = "harbor-scanner-adapter"
)

// Load configurations via viper.
func Load(configPath string) error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(fmt.Sprintf("/etc/%s/", app))
	viper.AddConfigPath(fmt.Sprintf("$HOME/.%s", app))
	viper.AddConfigPath(".")
	// Add extra config file.
	if len(strings.TrimSpace(configPath)) > 0 {
		viper.SetConfigFile(configPath)
	}

	// Set defaults.
	viper.SetDefault("server.protocol", "http")
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("scanner.workers", 5)

	// Read from env.
	viper.SetEnvPrefix("HSA")
	viper.AllowEmptyEnv(false)
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("Fatal error config file: %w \n", err)
	}

	return nil
}

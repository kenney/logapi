/**
 * Copyright 2015 Acquia, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package logapi - This library handles the file-based runtime configuration.
package logapi

import (
	"io/ioutil"
	"os"
	"regexp"

	"gopkg.in/yaml.v2"
)

// ConfigValues describes the data type that configuration is loaded into. The
// values from the YAML config file map directly to these values. e.g.
//
// service:
//     name: logapi
//     debug: false
//
// Map to:
// config.Service.Name = "logapi"
// config.Service.Debug = false
//
// All values specified in the ConfigValues struct should also have a default
// value set in LoadFile() to ensure a safe runtime environment.
type ConfigValues struct {
	Service struct {
		Name     string
		Hostname string
	}
	Connection struct {
		TCP struct {
			Enabled bool
			Host    string
			Port    int
		}
	}
	Redis struct {
		Host string
		Port int
		Auth string
	}
	Debug struct {
		Verbose bool
	}
}

// CreateConfig is a factory for creating ConfigValues.
func CreateConfig(filePath string) (ConfigValues, error) {
	config := new(ConfigValues)
	err := config.LoadFile(filePath)
	return *config, err
}

// LoadFile will read configuration from a specified file.
func (config *ConfigValues) LoadFile(filePath string) error {
	var err error

	// Establish all of the default values.

	// Service
	config.Service.Name = "logapi"
	config.Service.Hostname = ""

	// Connection
	config.Connection.TCP.Enabled = true
	config.Connection.TCP.Host = "127.0.0.1"
	config.Connection.TCP.Port = 8080
	
	// Redis
	config.Redis.Host = "127.0.0.1"
	config.Redis.Port = 6379
	config.Redis.Auth = ""
	
	// Debug
	config.Debug.Verbose = true

	// Attempt to read in the file.
	if filePath != "" {
		contents, readError := ioutil.ReadFile(filePath)
		if readError != nil {
			err = readError
		} else {
			err = yaml.Unmarshal([]byte(contents), &config)
		}
	}

	// If the hostname is empty, use the current one.
	config.Service.Hostname = GetHostname(config.Service.Hostname)
	return err
}

// GetHostname determines the current hostname if the provided default is empty.
func GetHostname(defaultValue string) (hostname string) {
	hostname = "unknown"
	if defaultValue == "" {
		hn, err := os.Hostname()
		if err == nil {
			hostname = hn
		}
	} else {
		hostname = defaultValue
	}
	re := regexp.MustCompile("[^a-zA-Z0-9]")
	hostname = re.ReplaceAllString(hostname, "-")

	return
}

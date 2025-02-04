// Copyright 2020 The Cloud Native Events Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package plugins

import (
	"fmt"
	"path/filepath"
	"plugin"
	"sync"

	"github.com/redhat-cne/cloud-event-proxy/pkg/common"
	log "github.com/sirupsen/logrus"

	v1amqp "github.com/redhat-cne/sdk-go/v1/amqp"
)

// Handler handler for loading plugins
type Handler struct {
	Path string
}

// LoadAMQPPlugin loads amqp plugin
func (pl Handler) LoadAMQPPlugin(wg *sync.WaitGroup, scConfig *common.SCConfiguration) (*v1amqp.AMQP, error) {
	log.Printf("Starting AMQP server")
	amqpPlugin, err := filepath.Glob(fmt.Sprintf("%s/amqp_plugin.so", pl.Path))
	if err != nil {
		log.Fatalf("cannot load amqp plugin %v\n", err)
		return nil, err
	}
	if len(amqpPlugin) == 0 {
		return nil, fmt.Errorf("amqp plugin not found in the path %s", pl.Path)
	}
	p, err := plugin.Open(amqpPlugin[0])
	if err != nil {
		log.Fatalf("cannot open amqp plugin %v", err)
		return nil, err
	}
	symbol, err := p.Lookup("Start")
	if err != nil {
		log.Fatalf("cannot open amqp plugin start method %v", err)
		return nil, err
	}

	startFunc, ok := symbol.(func(wg *sync.WaitGroup, scConfig *common.SCConfiguration) (*v1amqp.AMQP, error))
	if !ok {
		log.Fatalf("Plugin has no 'Start(*sync.WaitGroup,*common.SCConfiguration) (*v1_amqp.AMQP,error)' function")
		return nil, fmt.Errorf("plugin has no 'start(amqpHost string, dataIn <-chan channel.DataChan, dataOut chan<- channel.DataChan, close <-chan bool) (*v1_amqp.AMQP,error)' function")
	}
	amqpInstance, err := startFunc(wg, scConfig)
	if err != nil {
		log.Printf("error starting amqp at %s error: %v", scConfig.AMQPHost, err)
		return amqpInstance, err
	}
	return amqpInstance, nil
}

// LoadPTPPlugin loads ptp plugin
func (pl Handler) LoadPTPPlugin(wg *sync.WaitGroup, scConfig *common.SCConfiguration, fn func(e interface{}) error) error {
	ptpPlugin, err := filepath.Glob(fmt.Sprintf("%s/ptp_operator_plugin.so", pl.Path))
	if err != nil {
		log.Fatalf("cannot load ptp plugin %v", err)
	}
	if len(ptpPlugin) == 0 {
		return fmt.Errorf("ptp plugin not found in the path %s", pl.Path)
	}
	p, err := plugin.Open(ptpPlugin[0])
	if err != nil {
		log.Fatalf("cannot open ptp plugin %v", err)
		return err
	}

	symbol, err := p.Lookup("Start")
	if err != nil {
		log.Fatalf("cannot open ptp plugin start method %v", err)
		return err
	}
	startFunc, ok := symbol.(func(*sync.WaitGroup, *common.SCConfiguration, func(e interface{}) error) error)
	if !ok {
		log.Fatalf("Plugin has no 'Start(*sync.WaitGroup, *common.SCConfiguration,  fn func(e interface{}) error)(error)' function")
	}
	return startFunc(wg, scConfig, fn)
}

// LoadHwEventPlugin loads hw event plugin
func (pl Handler) LoadHwEventPlugin(wg *sync.WaitGroup, scConfig *common.SCConfiguration, fn func(e interface{}) error) error {
	hwPlugin, err := filepath.Glob(fmt.Sprintf("%s/hw_event_plugin.so", pl.Path))
	if err != nil {
		log.Fatalf("cannot load hw event plugin %v", err)
	}
	if len(hwPlugin) == 0 {
		return fmt.Errorf("hw event plugin not found in the path %s", pl.Path)
	}
	p, err := plugin.Open(hwPlugin[0])
	if err != nil {
		log.Fatalf("cannot open hw event plugin %v", err)
		return err
	}

	symbol, err := p.Lookup("Start")
	if err != nil {
		log.Fatalf("cannot open hw event plugin start method %v", err)
		return err
	}
	startFunc, ok := symbol.(func(*sync.WaitGroup, *common.SCConfiguration, func(e interface{}) error) error)
	if !ok {
		log.Fatalf("Plugin has no 'Start(*sync.WaitGroup, *common.SCConfiguration,  fn func(e interface{}) error)(error)' function")
	}
	return startFunc(wg, scConfig, fn)
}

// LoadMockPlugin loads mock test  plugin
func (pl Handler) LoadMockPlugin(wg *sync.WaitGroup, scConfig *common.SCConfiguration, fn func(e interface{}) error) error {
	mockPlugin, err := filepath.Glob(fmt.Sprintf("%s/mock_plugin.so", pl.Path))
	if err != nil {
		log.Fatalf("cannot load mock plugin %v", err)
	}
	if len(mockPlugin) == 0 {
		return fmt.Errorf("mock plugin not found in the path %s", pl.Path)
	}
	p, err := plugin.Open(mockPlugin[0])
	if err != nil {
		log.Fatalf("cannot open mock plugin %v", err)
		return err
	}

	symbol, err := p.Lookup("Start")
	if err != nil {
		log.Fatalf("cannot open mock plugin start method %v", err)
		return err
	}
	startFunc, ok := symbol.(func(*sync.WaitGroup, *common.SCConfiguration, func(e interface{}) error) error)
	if !ok {
		log.Fatalf("Plugin has no 'Start(*sync.WaitGroup, *common.SCConfiguration,  fn func(e interface{}) error)(error)' function")
	}
	return startFunc(wg, scConfig, fn)
}

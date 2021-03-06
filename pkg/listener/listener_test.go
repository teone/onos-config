// Copyright 2019-present Open Networking Foundation.
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

package listener

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/onosproject/onos-config/pkg/events"
	"github.com/onosproject/onos-config/pkg/southbound/topocache"
	"github.com/onosproject/onos-config/pkg/store"
	"gotest.tools/assert"
	is "gotest.tools/assert/cmp"
)

var (
	device1Channel, device2Channel, device3Channel chan events.ConfigEvent
	device1, device2, device3                      topocache.Device
	err                                            error
)

const (
	configStoreDefaultFileName = "../store/testout/configStore-sample.json"
	changeStoreDefaultFileName = "../store/testout/changeStore-sample.json"
)

var (
	configStoreTest store.ConfigurationStore
	changeStoreTest store.ChangeStore
)

func TestMain(m *testing.M) {
	device1 = topocache.Device{Addr: "localhost:10161"}
	device2 = topocache.Device{Addr: "localhost:10162"}
	device3 = topocache.Device{Addr: "localhost:10163"}

	device1Channel, err = Register(device1.Addr, true)
	device2Channel, err = Register(device2.Addr, true)
	device3Channel, err = Register(device3.Addr, true)

	configStoreTest, err = store.LoadConfigStore(configStoreDefaultFileName)
	if err != nil {
		wd, _ := os.Getwd()
		fmt.Println("Cannot load config store ", err, wd)
		return
	}
	fmt.Println("Configuration store loaded from", configStoreDefaultFileName)

	changeStoreTest, err = store.LoadChangeStore(changeStoreDefaultFileName)
	if err != nil {
		fmt.Println("Cannot load change store ", err)
		return
	}

	os.Exit(m.Run())
}

func Test_listlisteners(t *testing.T) {
	listeners := ListListeners()

	assert.Assert(t, len(listeners) == 3, "Expected to find 3 devices in list. Got %d", len(listeners))

	listenerStr := strings.Join(listeners, ",")
	// Could be in any order
	assert.Assert(t, is.Contains(listenerStr, "localhost:10161"), "Expected to find device1 in list. Got %s", listeners)
	assert.Assert(t, is.Contains(listenerStr, "localhost:10162"), "Expected to find device2 in list. Got %s", listeners)
	assert.Assert(t, is.Contains(listenerStr, "localhost:10163"), "Expected to find device3 in list. Got %s", listeners)
}

func Test_register(t *testing.T) {
	device4Channel, err := Register("device4", true)
	assert.NilError(t, err, "Unexpected error when registering device %s", err)

	var deviceChannelIf interface{} = device4Channel
	chanType, ok := deviceChannelIf.(chan events.ConfigEvent)
	assert.Assert(t, ok, "Unexpected channel type when registering device %v", chanType)

}

func Test_unregister(t *testing.T) {
	err1 := Unregister("device5", true)

	assert.Assert(t, err1 != nil, "Unexpected lack of error when unregistering non existent device")
	assert.Assert(t, is.Contains(err1.Error(), "had not been registered"), "Unexpected error text when unregistering non existent device %s", err1)

	_, err1 = Register("device6", true)
	_, err1 = Register("device7", true)

	err2 := Unregister("device6", true)
	assert.NilError(t, err2, "Unexpected error when unregistering device6 %s", err)
}

func Test_listen(t *testing.T) {
	// Start a test listener
	testChan := make(chan events.Event, 10)
	go testSync(testChan)
	// Start the main listener system
	changesChannel := make(chan events.ConfigEvent, 10)
	go Listen(changesChannel)
	changeID := []byte("test")
	// Send down some changes
	for i := 1; i < 13; i++ {
		event := events.CreateConfigEvent("device"+strconv.Itoa(i), changeID, true)

		changesChannel <- event
	}

	// Wait for the changes to get distributed
	time.Sleep(time.Second)
	Unregister(device1.Addr, true)
	Unregister(device2.Addr, true)
	Unregister(device3.Addr, true)
	close(testChan)
	close(changesChannel)
}

func testSync(testChan <-chan events.Event) {
	log.Println("Listen for config changes for Test")

	for nbiChange := range testChan {
		fmt.Println("Change for Test", nbiChange)
	}
}

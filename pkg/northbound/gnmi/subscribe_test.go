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

package gnmi

import (
	"context"
	"github.com/onosproject/onos-config/pkg/utils"
	"github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/grpc"
	"gotest.tools/assert"
	"log"
	"testing"
	"time"
)

type gNMISubscribeServerFake struct {
	Request   *gnmi.SubscribeRequest
	Responses chan *gnmi.SubscribeResponse
	grpc.ServerStream
}

func (x gNMISubscribeServerFake) Send(m *gnmi.SubscribeResponse) error {
	x.Responses <- m
	if m.GetSyncResponse() {
		close(x.Responses)
	}
	return nil
}

func (x gNMISubscribeServerFake) Recv() (*gnmi.SubscribeRequest, error) {
	return x.Request, nil
}

// Test_SubscribeLeafOnce tests subscribing with mode ONCE and then immediately receiving the subscription for a specific leaf.
func Test_SubscribeLeafOnce(t *testing.T) {
	server := setUp(false)

	path, err := utils.ParseGNMIElements([]string{"test1:cont1a", "cont2a", "leaf2a"})

	assert.NilError(t, err, "Unexpected error doing parsing")

	path.Target = "Device1"

	subscription := &gnmi.Subscription{
		Path: path,
		Mode: gnmi.SubscriptionMode_TARGET_DEFINED,
	}

	subscriptions := make([]*gnmi.Subscription, 0)

	subscriptions = append(subscriptions, subscription)

	subList := &gnmi.SubscriptionList{
		Subscription: subscriptions,
		Mode:         gnmi.SubscriptionList_ONCE,
	}

	request := &gnmi.SubscribeRequest{
		Request: &gnmi.SubscribeRequest_Subscribe{
			Subscribe: subList,
		},
	}

	responsesChan := make(chan *gnmi.SubscribeResponse)
	serverFake := gNMISubscribeServerFake{
		Request:   request,
		Responses: responsesChan,
	}
	go func() {
		err = server.Subscribe(serverFake)
	}()

	assert.NilError(t, err, "Unexpected error doing Subscribe")

	responseReq := <-responsesChan
	assert.Assert(t, responseReq.GetUpdate().Update != nil, "Update should not be nil")

	assert.Equal(t, len(responseReq.GetUpdate().Update), 1)

	if responseReq.GetUpdate().Update[0] == nil {
		log.Println("response should contain at least one update and that should not be nil")
		t.FailNow()
	}

	pathResponse := responseReq.GetUpdate().Update[0].Path

	assert.Equal(t, pathResponse.Target, "Device1")

	assert.Equal(t, len(pathResponse.Elem), 3, "Expected 3 path elements")

	assert.Equal(t, pathResponse.Elem[0].Name, "test1:cont1a")
	assert.Equal(t, pathResponse.Elem[1].Name, "cont2a")
	assert.Equal(t, pathResponse.Elem[2].Name, "leaf2a")

	assert.Equal(t, responseReq.GetUpdate().Update[0].Val.GetStringVal(), "13")
	close(responsesChan)

}

// Test_SubscribeLeafOnce tests subscribing with mode ONCE and then immediately receiving the subscription for a specific leaf.
func Test_SubscribeLeafStream(t *testing.T) {
	server := setUp(true)

	path, err := utils.ParseGNMIElements([]string{"cont1a", "cont2a", "leaf4a"})

	assert.NilError(t, err, "Unexpected error doing parsing")

	path.Target = "Device1"

	var deletePaths = make([]*gnmi.Path, 0)
	var replacedPaths = make([]*gnmi.Update, 0)
	var updatedPaths = make([]*gnmi.Update, 0)
	//augmenting pre-existing value by one
	typedValue := gnmi.TypedValue_StringVal{StringVal: "14"}
	value := gnmi.TypedValue{Value: &typedValue}
	updatedPaths = append(updatedPaths, &gnmi.Update{Path: path, Val: &value})

	var setRequest = gnmi.SetRequest{
		Delete:  deletePaths,
		Replace: replacedPaths,
		Update:  updatedPaths,
	}
	log.Println(path)
	subscription := &gnmi.Subscription{
		Path: path,
		Mode: gnmi.SubscriptionMode_TARGET_DEFINED,
	}

	subscriptions := make([]*gnmi.Subscription, 0)

	subscriptions = append(subscriptions, subscription)

	subList := &gnmi.SubscriptionList{
		Subscription: subscriptions,
		Mode:         gnmi.SubscriptionList_STREAM,
	}

	request := &gnmi.SubscribeRequest{
		Request: &gnmi.SubscribeRequest_Subscribe{
			Subscribe: subList,
		},
	}

	responsesChan := make(chan *gnmi.SubscribeResponse, 1)
	serverFake := gNMISubscribeServerFake{
		Request:   request,
		Responses: responsesChan,
	}

	go func() {
		err = server.Subscribe(serverFake)
	}()
	assert.NilError(t, err, "Unexpected error doing Subscribe")

	//FIXME Waiting for subscribe to finish properly --> when event is issued assuring state consistency we can remove
	time.Sleep(100000)

	_, err = server.Set(context.Background(), &setRequest)
	assert.NilError(t, err, "Unexpected error doing parsing")

	for response := range responsesChan {
		log.Println("response", response)
		if len(response.GetUpdate().GetUpdate()) != 0 {
			assert.Assert(t, response.GetUpdate().GetUpdate() != nil, "Update should not be nil")

			assert.Equal(t, len(response.GetUpdate().GetUpdate()), 1)

			if response.GetUpdate().GetUpdate()[0] == nil {
				log.Println("response should contain at least one update and that should not be nil")
				t.FailNow()
			}

			pathResponse := response.GetUpdate().GetUpdate()[0].Path

			assert.Equal(t, pathResponse.Target, "Device1")

			assert.Equal(t, len(pathResponse.Elem), 3, "Expected 3 path elements")

			assert.Equal(t, pathResponse.Elem[0].Name, "cont1a")
			assert.Equal(t, pathResponse.Elem[1].Name, "cont2a")
			assert.Equal(t, pathResponse.Elem[2].Name, "leaf4a")

			assert.Equal(t, response.GetUpdate().GetUpdate()[0].Val.GetStringVal(), "14")
		} else {
			assert.Equal(t, response.GetSyncResponse(), true, "Sync should be true")
		}
	}

}

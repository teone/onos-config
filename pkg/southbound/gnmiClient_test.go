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

package southbound

import (
	"context"
	"github.com/openconfig/gnmi/client"
	"gotest.tools/assert"
	"testing"
)

func Test_GnmiClientBadCreate(t *testing.T) {
	d := client.Destination{}
	ctx := context.Background()
	c, e := GnmiClientFactory(ctx, d)

	assert.ErrorContains(t, e, "Addrs must only contain")
	assert.Equal(t, c, nil)
}

// Copyright 2016-2020 terraform-provider-sakuracloud authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sakuracloud

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/search"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

func dataSourceSakuraCloudServerVNCInfo() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSakuraCloudServerVNCInfoRead,

		Schema: map[string]*schema.Schema{
			"server_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateSakuracloudIDType,
				Description:  "The id of the Server",
			},
			"host": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The host name for connecting by VNC",
			},
			"port": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The port number for connecting by VNC",
			},
			"password": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The password for connecting by VNC",
			},
			"zone": schemaDataSourceZone("Server VNC Information"),
		},
	}
}

func dataSourceSakuraCloudServerVNCInfoRead(d *schema.ResourceData, meta interface{}) error {
	client, zone, err := sakuraCloudClient(d, meta)
	if err != nil {
		return err
	}
	ctx, cancel := operationContext(d, schema.TimeoutRead)
	defer cancel()

	// validate account
	authOp := sacloud.NewAuthStatusOp(client)
	auth, err := authOp.Read(ctx)
	if err != nil {
		return fmt.Errorf("could not read Authentication Status: %s", err)
	}
	if auth.Permission == types.Permissions.View {
		return errors.New("current API key is only permitted to view")
	}

	// validate zone
	zoneOp := sacloud.NewZoneOp(client)
	searched, err := zoneOp.Find(ctx, &sacloud.FindCondition{
		Filter: search.Filter{
			search.Key("Name"): search.ExactMatch(zone),
		},
	})
	if err != nil || searched.Count == 0 {
		return fmt.Errorf("could not find SakuraCkoud Zone[%s]: %s", zone, err)
	}
	zoneInfo := searched.Zones[0]
	if zoneInfo.IsDummy {
		return fmt.Errorf("reading VNC information is failed: VNC information is not support on zone[%s]", zone)
	}

	serverOp := sacloud.NewServerOp(client)
	serverID := expandSakuraCloudID(d, "server_id")

	data, err := serverOp.GetVNCProxy(ctx, zone, serverID)
	if err != nil {
		return fmt.Errorf("could not get VNC information: %s", err)
	}

	d.SetId(serverID.String())
	d.Set("server_id", serverID.String()) // nolint
	d.Set("host", data.IOServerHost)      // nolint
	d.Set("port", data.Port.Int())        // nolint
	d.Set("password", data.Password)      // nolint
	d.Set("zone", zone)                   // nolint
	return nil
}

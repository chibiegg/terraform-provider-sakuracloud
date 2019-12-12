// Copyright 2016-2019 terraform-provider-sakuracloud authors
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
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

func flattenDiskPlan(data *sacloud.Disk) string {
	var plan string
	switch data.DiskPlanID {
	case types.DiskPlans.SSD:
		plan = "ssd"
	case types.DiskPlans.HDD:
		plan = "hdd"
	}
	return plan
}

func expandDiskPlan(d *schema.ResourceData) types.ID {
	var planID types.ID
	switch d.Get("plan").(string) {
	case "ssd":
		planID = types.DiskPlans.SSD
	case "hdd":
		planID = types.DiskPlans.HDD
	}
	return planID
}

func expandDiskCreateRequest(d *schema.ResourceData) *sacloud.DiskCreateRequest {

	return &sacloud.DiskCreateRequest{
		DiskPlanID:      expandDiskPlan(d),
		Connection:      types.EDiskConnection(d.Get("connector").(string)),
		SourceDiskID:    expandSakuraCloudID(d, "source_disk_id"),
		SourceArchiveID: expandSakuraCloudID(d, "source_archive_id"),
		SizeMB:          toSizeMB(d.Get("size").(int)),
		Name:            d.Get("name").(string),
		Description:     d.Get("description").(string),
		Tags:            expandTags(d),
		IconID:          expandSakuraCloudID(d, "icon_id"),
	}
}

func expandDiskUpdateRequest(d *schema.ResourceData) *sacloud.DiskUpdateRequest {
	return &sacloud.DiskUpdateRequest{
		Connection:  types.EDiskConnection(d.Get("connector").(string)),
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Tags:        expandTags(d),
		IconID:      expandSakuraCloudID(d, "icon_id"),
	}
}

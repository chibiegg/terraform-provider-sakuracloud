// Copyright 2016-2021 terraform-provider-sakuracloud authors
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
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/sacloud/libsacloud/v2/helper/query"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/ostype"
)

func dataSourceSakuraCloudArchive() *schema.Resource {
	resourceName := "Archive"

	return &schema.Resource{
		ReadContext: dataSourceSakuraCloudArchiveRead,

		Schema: map[string]*schema.Schema{
			filterAttrName: filterSchema(&filterSchemaOption{}),
			"os_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(ostype.OSTypeShortNames, false)),
				ConflictsWith:    []string{"filter"},
				Description: descf(
					"The criteria used to filter SakuraCloud archives. This must be one of following:  \n%s",
					ostype.OSTypeShortNames,
				),
			},
			"name":        schemaDataSourceName(resourceName),
			"size":        schemaDataSourceSize(resourceName),
			"icon_id":     schemaDataSourceIconID(resourceName),
			"description": schemaDataSourceDescription(resourceName),
			"tags":        schemaDataSourceTags(resourceName),
			"zone":        schemaDataSourceZone(resourceName),
		},
	}
}

func dataSourceSakuraCloudArchiveRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, zone, err := sakuraCloudClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	searcher := sacloud.NewArchiveOp(client)

	var data *sacloud.Archive
	if osType, ok := d.GetOk("os_type"); ok {
		strOSType := osType.(string)
		if strOSType != "" {
			res, err := query.FindArchiveByOSType(ctx, searcher, zone, ostype.StrToOSType(strOSType))
			if err != nil {
				return diag.FromErr(err)
			}
			data = res
		}
	} else {
		findCondition := &sacloud.FindCondition{}
		if rawFilter, ok := d.GetOk(filterAttrName); ok {
			findCondition.Filter = expandSearchFilter(rawFilter)
		}

		res, err := searcher.Find(ctx, zone, findCondition)
		if err != nil {
			return diag.Errorf("could not find SakuraCloud Archive resource: %s", err)
		}
		if res == nil || res.Count == 0 {
			return filterNoResultErr()
		}

		targets := res.Archives
		if len(targets) == 0 {
			return filterNoResultErr()
		}
		data = targets[0]
	}

	if data != nil {
		d.SetId(data.ID.String())
		d.Set("name", data.Name)               // nolint
		d.Set("size", data.GetSizeGB())        // nolint
		d.Set("icon_id", data.IconID.String()) // nolint
		d.Set("description", data.Description) // nolint
		if err := d.Set("tags", flattenTags(data.Tags)); err != nil {
			return diag.FromErr(err)
		}
		d.Set("zone", getZone(d, client)) // nolint
	}
	return nil
}

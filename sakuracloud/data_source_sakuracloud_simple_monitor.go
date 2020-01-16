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
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

func dataSourceSakuraCloudSimpleMonitor() *schema.Resource {
	resourceName := "SimpleMonitor"
	return &schema.Resource{
		Read: dataSourceSakuraCloudSimpleMonitorRead,

		Schema: map[string]*schema.Schema{
			filterAttrName: filterSchema(&filterSchemaOption{}),
			"target": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The monitoring target of the simple monitor. This will be IP address or FQDN",
			},
			"delay_loop": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The interval in seconds between checks",
			},
			"health_check": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"protocol": {
							Type:     schema.TypeString,
							Computed: true,
							Description: descf(
								"The protocol used for health checks. This will be one of [%s]",
								types.SimpleMonitorProtocolStrings,
							),
						},
						"host_header": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The value of host header send when checking by HTTP/HTTPS",
						},
						"path": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The path used when checking by HTTP/HTTPS",
						},
						"status": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The response-code to expect when checking by HTTP/HTTPS",
						},
						"sni": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "The flag to enable SNI when checking by HTTP/HTTPS",
						},
						"username": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The user name for basic auth used when checking by HTTP/HTTPS",
						},
						"password": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The password for basic auth used when checking by HTTP/HTTPS",
						},
						"port": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The target port number",
						},
						"qname": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The FQDN used when checking by DNS",
						},
						"excepcted_data": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The expected value used when checking by DNS",
						},
						"community": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The SNMP community string used when checking by SNMP",
						},
						"snmp_version": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The SNMP version used when checking by SNMP",
						},
						"oid": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The SNMP OID used when checking by SNMP",
						},
						"remaining_days": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The number of remaining days until certificate expiration used when checking SSL certificates",
						},
					},
				},
			},
			"icon_id":     schemaDataSourceIconID(resourceName),
			"description": schemaDataSourceDescription(resourceName),
			"tags":        schemaDataSourceTags(resourceName),
			"notify_email_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "The flag to enable notification by email",
			},
			"notify_email_html": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "The flag to enable HTML format instead of text format",
			},
			"notify_slack_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "The flag to enable notification by slack/discord",
			},
			"notify_slack_webhook": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The webhook URL for sending notification by slack/discord",
			},
			"notify_interval": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The interval in hours between notification",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "The flag to enable monitoring by the simple monitor",
			},
		},
	}
}

func dataSourceSakuraCloudSimpleMonitorRead(d *schema.ResourceData, meta interface{}) error {
	client, _, err := sakuraCloudClient(d, meta)
	if err != nil {
		return err
	}
	ctx, cancel := operationContext(d, schema.TimeoutRead)
	defer cancel()

	searcher := sacloud.NewSimpleMonitorOp(client)

	findCondition := &sacloud.FindCondition{}
	if rawFilter, ok := d.GetOk(filterAttrName); ok {
		findCondition.Filter = expandSearchFilter(rawFilter)
	}

	res, err := searcher.Find(ctx, findCondition)
	if err != nil {
		return fmt.Errorf("could not find SakuraCloud SimpleMonitor resource: %s", err)
	}
	if res == nil || res.Count == 0 || len(res.SimpleMonitors) == 0 {
		return filterNoResultErr()
	}

	targets := res.SimpleMonitors
	d.SetId(targets[0].ID.String())
	return setSimpleMonitorResourceData(ctx, d, client, targets[0])
}

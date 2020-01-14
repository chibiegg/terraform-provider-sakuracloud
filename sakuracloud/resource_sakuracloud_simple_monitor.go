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
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

func resourceSakuraCloudSimpleMonitor() *schema.Resource {
	resourceName := "SimpleMonitor"

	return &schema.Resource{
		Create: resourceSakuraCloudSimpleMonitorCreate,
		Read:   resourceSakuraCloudSimpleMonitorRead,
		Update: resourceSakuraCloudSimpleMonitorUpdate,
		Delete: resourceSakuraCloudSimpleMonitorDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"target": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The monitoring target of the simple monitor. This must be IP address or FQDN",
			},
			"delay_loop": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(60, 3600),
				Default:      60,
				Description: descf(
					"The interval in seconds between checks. %s",
					descRange(60, 3600),
				),
			},
			"health_check": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"protocol": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(types.SimpleMonitorProtocolsStrings, false),
							Description: descf(
								"The protocol used for health checks. This must be one of [%s]",
								types.SimpleMonitorProtocolsStrings,
							),
						},
						"host_header": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The value of host header send when checking by HTTP/HTTPS",
						},
						"path": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The path used when checking by HTTP/HTTPS",
						},
						"status": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "The response-code to expect when checking by HTTP/HTTPS",
						},
						"sni": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "The flag to enable SNI when checking by HTTP/HTTPS",
						},
						"username": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The user name for basic auth used when checking by HTTP/HTTPS",
						},
						"password": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The password for basic auth used when checking by HTTP/HTTPS",
						},
						"port": {
							Type:        schema.TypeInt,
							Optional:    true,
							Computed:    true,
							Description: "The target port number",
						},
						"qname": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The FQDN used when checking by DNS",
						},
						"excepcted_data": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The expected value used when checking by DNS",
						},
						"community": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The SNMP community string used when checking by SNMP",
						},
						"snmp_version": {
							Type:         schema.TypeString,
							ValidateFunc: validation.StringInSlice([]string{"1", "2c"}, false),
							Optional:     true,
							Description: descf(
								"The SNMP version used when checking by SNMP. This must be one of %s",
								[]string{"1", "2c"},
							),
						},
						"oid": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The SNMP OID used when checking by SNMP",
						},
						"remaining_days": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 9999),
							Default:      30,
							Description: descf(
								"The number of remaining days until certificate expiration used when checking SSL certificates. %s",
								descRange(1, 9999),
							),
						},
					},
				},
			},
			"icon_id":     schemaResourceIconID(resourceName),
			"description": schemaResourceDescription(resourceName),
			"tags":        schemaResourceTags(resourceName),
			"notify_email_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "The flag to enable notification by email",
			},
			"notify_email_html": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "The flag to enable HTML format instead of text format",
			},
			"notify_slack_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "The flag to enable notification by slack/discord",
			},
			"notify_slack_webhook": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The webhook URL for sending notification by slack/discord",
			},
			"notify_interval": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     2,
				Description: "The interval in hours between notification",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "The flag to enable monitoring by the simple monitor",
			},
		},
	}
}

func resourceSakuraCloudSimpleMonitorCreate(d *schema.ResourceData, meta interface{}) error {
	client, _, err := sakuraCloudClient(d, meta)
	if err != nil {
		return err
	}
	ctx, cancel := operationContext(d, schema.TimeoutCreate)
	defer cancel()

	smOp := sacloud.NewSimpleMonitorOp(client)

	simpleMonitor, err := smOp.Create(ctx, expandSimpleMonitorCreateRequest(d))
	if err != nil {
		return fmt.Errorf("creating SimpleMonitor is failed: %s", err)
	}

	d.SetId(simpleMonitor.ID.String())
	return resourceSakuraCloudSimpleMonitorRead(d, meta)
}

func resourceSakuraCloudSimpleMonitorRead(d *schema.ResourceData, meta interface{}) error {
	client, _, err := sakuraCloudClient(d, meta)
	if err != nil {
		return err
	}
	ctx, cancel := operationContext(d, schema.TimeoutRead)
	defer cancel()

	smOp := sacloud.NewSimpleMonitorOp(client)

	simpleMonitor, err := smOp.Read(ctx, sakuraCloudID(d.Id()))
	if err != nil {
		if sacloud.IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("could not read SimpleMonitor[%s]: %s", d.Id(), err)
	}

	return setSimpleMonitorResourceData(ctx, d, client, simpleMonitor)
}

func resourceSakuraCloudSimpleMonitorUpdate(d *schema.ResourceData, meta interface{}) error {
	client, _, err := sakuraCloudClient(d, meta)
	if err != nil {
		return err
	}
	ctx, cancel := operationContext(d, schema.TimeoutUpdate)
	defer cancel()

	smOp := sacloud.NewSimpleMonitorOp(client)

	simpleMonitor, err := smOp.Read(ctx, sakuraCloudID(d.Id()))
	if err != nil {
		return fmt.Errorf("could not read SimpleMonitor[%s]: %s", d.Id(), err)
	}

	simpleMonitor, err = smOp.Update(ctx, simpleMonitor.ID, expandSimpleMonitorUpdateRequest(d))
	if err != nil {
		return fmt.Errorf("updating SimpleMonitor[%s] is failed: %s", simpleMonitor.ID, err)
	}

	return resourceSakuraCloudSimpleMonitorRead(d, meta)
}

func resourceSakuraCloudSimpleMonitorDelete(d *schema.ResourceData, meta interface{}) error {
	client, _, err := sakuraCloudClient(d, meta)
	if err != nil {
		return err
	}
	ctx, cancel := operationContext(d, schema.TimeoutDelete)
	defer cancel()

	smOp := sacloud.NewSimpleMonitorOp(client)

	simpleMonitor, err := smOp.Read(ctx, sakuraCloudID(d.Id()))
	if err != nil {
		if sacloud.IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("could not read SimpleMonitor[%s]: %s", d.Id(), err)
	}

	if err := smOp.Delete(ctx, simpleMonitor.ID); err != nil {
		return fmt.Errorf("deleting SimpleMonitor[%s] is failed: %s", simpleMonitor.ID, err)
	}
	return nil
}

func setSimpleMonitorResourceData(ctx context.Context, d *schema.ResourceData, client *APIClient, data *sacloud.SimpleMonitor) error {
	d.Set("target", data.Target)                                       // nolint
	d.Set("delay_loop", data.DelayLoop)                                // nolint
	d.Set("icon_id", data.IconID.String())                             // nolint
	d.Set("description", data.Description)                             // nolint
	d.Set("enabled", data.Enabled.Bool())                              // nolint
	d.Set("notify_email_enabled", data.NotifyEmailEnabled.Bool())      // nolint
	d.Set("notify_email_html", data.NotifyEmailHTML.Bool())            // nolint
	d.Set("notify_slack_enabled", data.NotifySlackEnabled.Bool())      // nolint
	d.Set("notify_slack_webhook", data.SlackWebhooksURL)               // nolint
	d.Set("notify_interval", flattenSimpleMonitorNotifyInterval(data)) // nolint
	if err := d.Set("health_check", flattenSimpleMonitorHealthCheck(data)); err != nil {
		return err
	}
	return d.Set("tags", flattenTags(data.Tags))
}

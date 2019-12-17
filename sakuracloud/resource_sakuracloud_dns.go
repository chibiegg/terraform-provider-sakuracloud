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

func resourceSakuraCloudDNS() *schema.Resource {
	return &schema.Resource{
		Create: resourceSakuraCloudDNSCreate,
		Read:   resourceSakuraCloudDNSRead,
		Update: resourceSakuraCloudDNSUpdate,
		Delete: resourceSakuraCloudDNSDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		CustomizeDiff: hasTagResourceCustomizeDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"dns_servers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"record": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1000,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(types.DNSRecordTypesStrings(), false),
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
						"ttl": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  defaultTTL,
						},
						"priority": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 65535),
						},
						"weight": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 65535),
						},
						"port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 65535),
						},
					},
				},
			},
			"icon_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateSakuracloudIDType,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceSakuraCloudDNSCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*APIClient)
	ctx, cancel := operationContext(d, schema.TimeoutCreate)
	defer cancel()

	dnsOp := sacloud.NewDNSOp(client)

	dns, err := dnsOp.Create(ctx, expandDNSCreateRequest(d))
	if err != nil {
		return fmt.Errorf("creating SakuraCloud DNS is failed: %s", err)
	}

	d.SetId(dns.ID.String())
	return resourceSakuraCloudDNSRead(d, meta)
}

func resourceSakuraCloudDNSRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*APIClient)
	ctx, cancel := operationContext(d, schema.TimeoutRead)
	defer cancel()

	dnsOp := sacloud.NewDNSOp(client)

	dns, err := dnsOp.Read(ctx, sakuraCloudID(d.Id()))
	if err != nil {
		if sacloud.IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("could not read SakuraCloud DNS[%s]: %s", d.Id(), err)
	}

	return setDNSResourceData(ctx, d, client, dns)
}

func resourceSakuraCloudDNSUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*APIClient)
	ctx, cancel := operationContext(d, schema.TimeoutUpdate)
	defer cancel()

	dnsOp := sacloud.NewDNSOp(client)

	sakuraMutexKV.Lock(d.Id())
	defer sakuraMutexKV.Unlock(d.Id())

	dns, err := dnsOp.Read(ctx, sakuraCloudID(d.Id()))
	if err != nil {
		return fmt.Errorf("could not read SakuraCloud DNS[%s]: %s", d.Id(), err)
	}

	if _, err := dnsOp.Update(ctx, dns.ID, expandDNSUpdateRequest(d, dns)); err != nil {
		return fmt.Errorf("updating SakuraCloud DNS[%s] is failed: %s", d.Id(), err)
	}
	return resourceSakuraCloudDNSRead(d, meta)
}

func resourceSakuraCloudDNSDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*APIClient)
	ctx, cancel := operationContext(d, schema.TimeoutDelete)
	defer cancel()

	dnsOp := sacloud.NewDNSOp(client)

	sakuraMutexKV.Lock(d.Id())
	defer sakuraMutexKV.Unlock(d.Id())

	dns, err := dnsOp.Read(ctx, sakuraCloudID(d.Id()))
	if err != nil {
		if sacloud.IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("could not read SakuraCloud DNS[%s]: %s", d.Id(), err)
	}

	if err := dnsOp.Delete(ctx, dns.ID); err != nil {
		return fmt.Errorf("deleting SakuraCloud DNS[%s] is failed: %s", d.Id(), err)
	}
	return nil
}

func setDNSResourceData(ctx context.Context, d *schema.ResourceData, client *APIClient, data *sacloud.DNS) error {
	d.Set("zone", data.Name)               // nolint
	d.Set("icon_id", data.IconID.String()) // nolint
	d.Set("description", data.Description) // nolint
	if err := d.Set("dns_servers", data.DNSNameServers); err != nil {
		return err
	}
	if err := d.Set("record", flattenDNSRecords(data)); err != nil {
		return err
	}
	return d.Set("tags", data.Tags)
}

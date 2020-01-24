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
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/accessor"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
	"github.com/sacloud/libsacloud/v2/utils/power"
	"github.com/sacloud/libsacloud/v2/utils/setup"
)

func resourceSakuraCloudNFS() *schema.Resource {
	resourceName := "NFS"
	return &schema.Resource{
		Create: resourceSakuraCloudNFSCreate,
		Read:   resourceSakuraCloudNFSRead,
		Update: resourceSakuraCloudNFSUpdate,
		Delete: resourceSakuraCloudNFSDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(24 * time.Hour),
			Update: schema.DefaultTimeout(24 * time.Hour),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": schemaResourceName(resourceName),
			"plan": schemaResourcePlan(resourceName, "hdd", types.NFSPlanStrings),
			"size": schemaResourceSize(resourceName, 100),
			"network_interface": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"switch_id": schemaResourceSwitchID(resourceName),
						"ip_address": {
							Type:        schema.TypeString,
							ForceNew:    true,
							Required:    true,
							Description: descf("The IP address to assign to the %s", resourceName),
						},
						"netmask": {
							Type:         schema.TypeInt,
							ForceNew:     true,
							Required:     true,
							ValidateFunc: validation.IntBetween(8, 29),
							Description: descf(
								"The bit length of the subnet to assign to the %s. %s",
								resourceName,
								descRange(8, 29),
							),
						},
						"gateway": {
							Type:        schema.TypeString,
							ForceNew:    true,
							Optional:    true,
							Description: descf("The IP address of the gateway used by %s", resourceName),
						},
					},
				},
			},
			"icon_id":     schemaResourceIconID(resourceName),
			"description": schemaResourceDescription(resourceName),
			"tags":        schemaResourceTags(resourceName),
			"zone":        schemaResourceZone(resourceName),
		},
	}
}

func resourceSakuraCloudNFSCreate(d *schema.ResourceData, meta interface{}) error {
	client, zone, err := sakuraCloudClient(d, meta)
	if err != nil {
		return err
	}
	ctx, cancel := operationContext(d, schema.TimeoutCreate)
	defer cancel()

	nfsOp := sacloud.NewNFSOp(client)

	planID, err := expandNFSDiskPlanID(ctx, client, d)
	if err != nil {
		return fmt.Errorf("finding NFS plans is failed: %s", err)
	}

	builder := &setup.RetryableSetup{
		Create: func(ctx context.Context, zone string) (accessor.ID, error) {
			return nfsOp.Create(ctx, zone, expandNFSCreateRequest(d, planID))
		},
		Delete: func(ctx context.Context, zone string, id types.ID) error {
			return nfsOp.Delete(ctx, zone, id)
		},
		Read: func(ctx context.Context, zone string, id types.ID) (interface{}, error) {
			return nfsOp.Read(ctx, zone, id)
		},
		RetryCount:    3,
		IsWaitForCopy: true,
		IsWaitForUp:   true,
	}

	res, err := builder.Setup(ctx, zone)
	if err != nil {
		return fmt.Errorf("creating SakuraCloud NFS is failed: %s", err)
	}

	nfs, ok := res.(*sacloud.NFS)
	if !ok {
		return errors.New("creating SakuraCloud NFS is failed: created resource is not *sacloud.NFS")
	}

	d.SetId(nfs.ID.String())
	return resourceSakuraCloudNFSRead(d, meta)
}

func resourceSakuraCloudNFSRead(d *schema.ResourceData, meta interface{}) error {
	client, zone, err := sakuraCloudClient(d, meta)
	if err != nil {
		return err
	}
	ctx, cancel := operationContext(d, schema.TimeoutRead)
	defer cancel()

	nfsOp := sacloud.NewNFSOp(client)

	nfs, err := nfsOp.Read(ctx, zone, sakuraCloudID(d.Id()))
	if err != nil {
		if sacloud.IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("could not read SakuraCloud NFS[%s]: %s", d.Id(), err)
	}

	return setNFSResourceData(ctx, d, client, nfs)
}

func resourceSakuraCloudNFSUpdate(d *schema.ResourceData, meta interface{}) error {
	client, zone, err := sakuraCloudClient(d, meta)
	if err != nil {
		return err
	}
	ctx, cancel := operationContext(d, schema.TimeoutUpdate)
	defer cancel()

	nfsOp := sacloud.NewNFSOp(client)

	nfs, err := nfsOp.Read(ctx, zone, sakuraCloudID(d.Id()))
	if err != nil {
		return fmt.Errorf("could not read SakuraCloud NFS[%s]: %s", d.Id(), err)
	}

	_, err = nfsOp.Update(ctx, zone, nfs.ID, expandNFSUpdateRequest(d))
	if err != nil {
		return fmt.Errorf("updating SakuraCloud NFS[%s] is failed: %s", d.Id(), err)
	}

	return resourceSakuraCloudNFSRead(d, meta)
}

func resourceSakuraCloudNFSDelete(d *schema.ResourceData, meta interface{}) error {
	client, zone, err := sakuraCloudClient(d, meta)
	if err != nil {
		return err
	}
	ctx, cancel := operationContext(d, schema.TimeoutDelete)
	defer cancel()

	nfsOp := sacloud.NewNFSOp(client)

	nfs, err := nfsOp.Read(ctx, zone, sakuraCloudID(d.Id()))
	if err != nil {
		if sacloud.IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("could not read SakuraCloud NFS[%s]: %s", d.Id(), err)
	}

	if err := power.ShutdownNFS(ctx, nfsOp, zone, nfs.ID, true); err != nil {
		return err
	}

	if err := nfsOp.Delete(ctx, zone, nfs.ID); err != nil {
		return fmt.Errorf("deleting SakuraCloud NFS[%s] is failed: %s", d.Id(), err)
	}

	return nil
}

func setNFSResourceData(ctx context.Context, d *schema.ResourceData, client *APIClient, data *sacloud.NFS) error {
	if data.Availability.IsFailed() {
		d.SetId("")
		return fmt.Errorf("got unexpected state: NFS[%d].Availability is failed", data.ID)
	}

	plan, size, err := flattenNFSDiskPlan(ctx, client, data.PlanID)
	if err != nil {
		return err
	}
	d.Set("plan", plan) // nolint
	d.Set("size", size) // nolint
	if err := d.Set("network_interface", flattenNFSNetworkInterface(data)); err != nil {
		return err
	}
	d.Set("name", data.Name)               // nolint
	d.Set("icon_id", data.IconID.String()) // nolint
	d.Set("description", data.Description) // nolint
	d.Set("zone", getZone(d, client))      // nolint
	return d.Set("tags", flattenTags(data.Tags))
}

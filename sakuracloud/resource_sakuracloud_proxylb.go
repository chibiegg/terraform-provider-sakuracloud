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

func resourceSakuraCloudProxyLB() *schema.Resource {
	resourceName := "ProxyLB"

	return &schema.Resource{
		Create: resourceSakuraCloudProxyLBCreate,
		Read:   resourceSakuraCloudProxyLBRead,
		Update: resourceSakuraCloudProxyLBUpdate,
		Delete: resourceSakuraCloudProxyLBDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": schemaResourceName(resourceName),
			"plan": schemaResourceIntPlan(resourceName, types.ProxyLBPlans.CPS100.Int(), types.ProxyLBPlanValues),
			"vip_failover": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Description: "The flag to enable VIP fail-over",
			},
			"sticky_session": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "The flag to enable sticky session",
			},
			"timeout": {
				Type:        schema.TypeInt,
				Default:     10,
				Optional:    true,
				Description: "The timeout duration in seconds",
			},
			"region": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      types.ProxyLBRegions.IS1.String(),
				ValidateFunc: validation.StringInSlice(types.ProxyLBRegionStrings, false),
				ForceNew:     true,
				Description: descf(
					"The name of region that the proxy LB is in. This must be one of [%s]",
					types.ProxyLBRegionStrings,
				),
			},
			"bind_port": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 2,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"proxy_mode": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(types.ProxyLBProxyModeStrings, false),
							Description: descf(
								"The proxy mode. This must be one of [%s]",
								types.ProxyLBProxyModeStrings,
							),
						},
						"port": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "The number of listening port",
						},
						"redirect_to_https": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "The flag to enable redirection from http to https. This flag is used only when `proxy_mode` is `http`",
						},
						"support_http2": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "The flag to enable HTTP/2. This flag is used only when `proxy_mode` is `https`",
						},
						"response_header": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 10,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"header": {
										Type:        schema.TypeString,
										Required:    true,
										Description: descf("The field name of HTTP header added to response by the %s", resourceName),
									},
									"value": {
										Type:        schema.TypeString,
										Required:    true,
										Description: descf("The field value of HTTP header added to response by the %s", resourceName),
									},
								},
							},
						},
					},
				},
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
							ValidateFunc: validation.StringInSlice(types.ProxyLBProtocolStrings, false),
							Description: descf(
								"The protocol used for health checks. This must be one of [%s]",
								types.ProxyLBProtocolStrings,
							),
						},
						"delay_loop": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(10, 60),
							Default:      10,
							Description: descf(
								"The interval in seconds between checks. %s",
								descRange(10, 60),
							),
						},
						"host_header": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The value of host header send when checking by HTTP",
						},
						"path": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The path used when checking by HTTP",
						},
						"port": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "The port number used when checking by TCP",
						},
					},
				},
			},
			"sorry_server": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_address": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The IP address of the SorryServer. This will be used when all servers are down",
						},
						"port": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "The port number of the SorryServer. This will be used when all servers are down",
						},
					},
				},
			},
			"certificate": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"server_cert": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "The certificate for a server",
						},
						"intermediate_cert": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "The intermediate certificate for a server",
						},
						"private_key": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "The private key for a server",
						},
						"additional_certificate": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 19,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"server_cert": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The certificate for a server",
									},
									"intermediate_cert": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "The intermediate certificate for a server",
									},
									"private_key": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The private key for a server",
									},
								},
							},
						},
					},
				},
			},
			"server": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 40,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_address": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The IP address of the destination server",
						},
						"port": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 65535),
							Description:  descf("The port number of the destination server. %s", descRange(1, 65535)),
						},
						"group": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 10),
							Description: descf(
								"The name of load balancing group. This is used when using rule-based load balancing. %s",
								descLength(1, 10),
							),
						},
						"enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "The flag to enable as destination of load balancing",
						},
					},
				},
			},
			"rule": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The value of HTTP host header that is used as condition of rule-based balancing",
						},
						"path": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The request path that is used as condition of rule-based balancing",
						},
						"group": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 10),
							Description: descf(
								"The name of load balancing group. When proxyLB received request which matched to `host` and `path`, proxyLB forwards the request to servers that having same group name. %s",
								descLength(1, 10),
							),
						},
					},
				},
			},
			"icon_id":     schemaResourceIconID(resourceName),
			"description": schemaResourceDescription(resourceName),
			"tags":        schemaResourceTags(resourceName),
			"fqdn": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: descf("The FQDN for accessing to the %s. This is typically used as value of CNAME record", resourceName),
			},
			"vip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: descf("The virtual IP address assigned to the %s", resourceName),
			},
			"proxy_networks": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: descf("A list of CIDR block used by the %s to access the server", resourceName),
			},
		},
	}
}

func resourceSakuraCloudProxyLBCreate(d *schema.ResourceData, meta interface{}) error {
	client, _, err := sakuraCloudClient(d, meta)
	if err != nil {
		return err
	}
	ctx, cancel := operationContext(d, schema.TimeoutCreate)
	defer cancel()

	proxyLBOp := sacloud.NewProxyLBOp(client)

	proxyLB, err := proxyLBOp.Create(ctx, expandProxyLBCreateRequest(d))
	if err != nil {
		return fmt.Errorf("creating SakuraCloud ProxyLB is failed: %s", err)
	}

	certs := expandProxyLBCerts(d)
	if certs != nil {
		_, err := proxyLBOp.SetCertificates(ctx, proxyLB.ID, &sacloud.ProxyLBSetCertificatesRequest{
			PrimaryCerts:    certs.PrimaryCert,
			AdditionalCerts: certs.AdditionalCerts,
		})
		if err != nil {
			return fmt.Errorf("setting Certificates to ProxyLB[%s] is failed: %s", proxyLB.ID, err)
		}
	}

	d.SetId(proxyLB.ID.String())
	return resourceSakuraCloudProxyLBRead(d, meta)
}

func resourceSakuraCloudProxyLBRead(d *schema.ResourceData, meta interface{}) error {
	client, _, err := sakuraCloudClient(d, meta)
	if err != nil {
		return err
	}
	ctx, cancel := operationContext(d, schema.TimeoutRead)
	defer cancel()

	proxyLBOp := sacloud.NewProxyLBOp(client)

	proxyLB, err := proxyLBOp.Read(ctx, sakuraCloudID(d.Id()))
	if err != nil {
		if sacloud.IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("could not read SakuraCloud ProxyLB[%s]: %s", d.Id(), err)
	}

	return setProxyLBResourceData(ctx, d, client, proxyLB)
}

func resourceSakuraCloudProxyLBUpdate(d *schema.ResourceData, meta interface{}) error {
	client, _, err := sakuraCloudClient(d, meta)
	if err != nil {
		return err
	}
	ctx, cancel := operationContext(d, schema.TimeoutUpdate)
	defer cancel()

	proxyLBOp := sacloud.NewProxyLBOp(client)

	sakuraMutexKV.Lock(d.Id())
	defer sakuraMutexKV.Unlock(d.Id())

	proxyLB, err := proxyLBOp.Read(ctx, sakuraCloudID(d.Id()))
	if err != nil {
		return fmt.Errorf("could not read SakuraCloud ProxyLB[%s]: %s", d.Id(), err)
	}

	proxyLB, err = proxyLBOp.Update(ctx, proxyLB.ID, expandProxyLBUpdateRequest(d))
	if err != nil {
		return fmt.Errorf("updating SakuraCloud ProxyLB[%s] is failed: %s", d.Id(), err)
	}

	if d.HasChange("plan") {
		newPlan := types.EProxyLBPlan(d.Get("plan").(int))
		upd, err := proxyLBOp.ChangePlan(ctx, proxyLB.ID, &sacloud.ProxyLBChangePlanRequest{Plan: newPlan})
		if err != nil {
			return fmt.Errorf("changing ProxyLB[%s] plan is failed: %s", d.Id(), err)
		}

		// update ID
		proxyLB = upd
		d.SetId(proxyLB.ID.String())
	}

	if proxyLB.LetsEncrypt == nil && d.HasChange("certificate") {
		certs := expandProxyLBCerts(d)
		if certs == nil {
			if err := proxyLBOp.DeleteCertificates(ctx, proxyLB.ID); err != nil {
				return fmt.Errorf("deleting Certificates of ProxyLB[%s] is failed: %s", d.Id(), err)
			}
		} else {
			if _, err := proxyLBOp.SetCertificates(ctx, proxyLB.ID, &sacloud.ProxyLBSetCertificatesRequest{
				PrimaryCerts:    certs.PrimaryCert,
				AdditionalCerts: certs.AdditionalCerts,
			}); err != nil {
				return fmt.Errorf("setting Certificates to ProxyLB[%s] is failed: %s", d.Id(), err)
			}
		}
	}
	return resourceSakuraCloudProxyLBRead(d, meta)
}

func resourceSakuraCloudProxyLBDelete(d *schema.ResourceData, meta interface{}) error {
	client, _, err := sakuraCloudClient(d, meta)
	if err != nil {
		return err
	}
	ctx, cancel := operationContext(d, schema.TimeoutDelete)
	defer cancel()

	proxyLBOp := sacloud.NewProxyLBOp(client)

	sakuraMutexKV.Lock(d.Id())
	defer sakuraMutexKV.Unlock(d.Id())

	proxyLB, err := proxyLBOp.Read(ctx, sakuraCloudID(d.Id()))
	if err != nil {
		if sacloud.IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("could not read SakuraCloud ProxyLB[%s]: %s", d.Id(), err)
	}

	if err := proxyLBOp.Delete(ctx, proxyLB.ID); err != nil {
		return fmt.Errorf("deleting ProxyLB[%s] is failed: %s", d.Id(), err)
	}
	return nil
}

func setProxyLBResourceData(ctx context.Context, d *schema.ResourceData, client *APIClient, data *sacloud.ProxyLB) error {
	// certificates
	proxyLBOp := sacloud.NewProxyLBOp(client)

	certs, err := proxyLBOp.GetCertificates(ctx, data.ID)
	if err != nil {
		// even if certificate is deleted, it will not result in an error
		return err
	}

	d.Set("name", data.Name)                                   // nolint
	d.Set("plan", data.Plan.Int())                             // nolint
	d.Set("vip_failover", data.UseVIPFailover)                 // nolint
	d.Set("sticky_session", flattenProxyLBStickySession(data)) // nolint
	d.Set("timeout", flattenProxyLBTimeout(data))              // nolint
	d.Set("region", data.Region.String())                      // nolint
	d.Set("fqdn", data.FQDN)                                   // nolint
	d.Set("vip", data.VirtualIPAddress)                        // nolint
	d.Set("proxy_networks", data.ProxyNetworks)                // nolint
	d.Set("icon_id", data.IconID.String())                     // nolint
	d.Set("description", data.Description)                     // nolint
	if err := d.Set("bind_port", flattenProxyLBBindPorts(data)); err != nil {
		return err
	}
	if err := d.Set("health_check", flattenProxyLBHealthCheck(data)); err != nil {
		return err
	}
	if err := d.Set("sorry_server", flattenProxyLBSorryServer(data)); err != nil {
		return err
	}
	if err := d.Set("server", flattenProxyLBServers(data)); err != nil {
		return err
	}
	if err := d.Set("rule", flattenProxyLBRules(data)); err != nil {
		return err
	}
	if err := d.Set("certificate", flattenProxyLBCerts(certs)); err != nil {
		return err
	}
	return d.Set("tags", flattenTags(data.Tags))
}

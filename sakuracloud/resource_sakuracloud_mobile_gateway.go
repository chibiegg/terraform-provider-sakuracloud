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
	"math"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/sacloud/libsacloud/api"
	"github.com/sacloud/libsacloud/sacloud"
)

func resourceSakuraCloudMobileGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceSakuraCloudMobileGatewayCreate,
		Read:   resourceSakuraCloudMobileGatewayRead,
		Update: resourceSakuraCloudMobileGatewayUpdate,
		Delete: resourceSakuraCloudMobileGatewayDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		CustomizeDiff: hasTagResourceCustomizeDiff,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"switch_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateSakuracloudIDType,
			},
			"public_ipaddress": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_nw_mask_len": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"private_ipaddress": {
				Type:         schema.TypeString,
				ValidateFunc: validateIPv4Address(),
				Optional:     true,
			},
			"private_nw_mask_len": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(8, 29),
			},
			"internet_connection": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"dns_server1": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateIPv4Address(),
			},
			"dns_server2": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateIPv4Address(),
			},
			"traffic_control": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"quota": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, math.MaxInt32),
						},
						"band_width_limit": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, math.MaxInt32),
						},
						"enable_email": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"enable_slack": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"slack_webhook": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`^https://hooks.slack.com/services/\w+/\w+/\w+$`), ""),
						},
						"auto_traffic_shaping": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"static_route": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"prefix": {
							Type:     schema.TypeString,
							Required: true,
						},
						"next_hop": {
							Type:     schema.TypeString,
							Required: true,
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
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"sim_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			powerManageTimeoutKey: powerManageTimeoutParam,
			"zone": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				Description:  "target SakuraCloud zone",
				ValidateFunc: validateZone([]string{"is1a", "is1b", "tk1a", "tk1v"}),
			},
		},
	}
}

func resourceSakuraCloudMobileGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	client := getSacloudAPIClient(d, meta)

	var switchID int64
	var ip string
	var nwMaskLen int
	if rawSwitchID, ok := d.GetOk("switch_id"); ok {
		strSwitchID := rawSwitchID.(string)
		if strSwitchID != "" {
			switchID = toSakuraCloudID(strSwitchID).Int64()
			if rawIP, ok := d.GetOk("private_ipaddress"); ok {
				ip = rawIP.(string)
			}
			if rawNWMaskLen, ok := d.GetOk("private_nw_mask_len"); ok {
				nwMaskLen = rawNWMaskLen.(int)
			}

			if ip == "" || nwMaskLen == 0 {
				return errors.New("MobileGateway needs private_ipaddress and private_nw_mask_len when switch_id is specified")
			}
		}
	}

	opts := &sacloud.CreateMobileGatewayValue{}
	opts.Name = d.Get("name").(string)
	if iconID, ok := d.GetOk("icon_id"); ok {
		opts.IconID = toSakuraCloudID(iconID.(string))
	}
	if description, ok := d.GetOk("description"); ok {
		opts.Description = description.(string)
	}
	if rawTags, ok := d.GetOk("tags"); ok {
		if rawTags != nil {
			opts.Tags = expandTags(client, rawTags.([]interface{}))
		}
	}

	setting := &sacloud.MobileGatewaySetting{}
	setting.InternetConnection = &sacloud.MGWInternetConnection{
		Enabled: "False",
	}
	if d.Get("internet_connection").(bool) {
		setting.InternetConnection.Enabled = "True"
	}

	createMgw, err := sacloud.CreateNewMobileGateway(opts, setting)
	if err != nil {
		return fmt.Errorf("Failed to create SakuraCloud MobileGateway resource: %s", err)
	}

	mgw, err := client.MobileGateway.Create(createMgw)
	if err != nil {
		return fmt.Errorf("Failed to create SakuraCloud MobileGateway resource: %s", err)
	}

	//wait
	err = client.MobileGateway.SleepWhileCopying(mgw.ID, client.DefaultTimeoutDuration, 20)
	if err != nil {
		return fmt.Errorf("Failed to create SakuraCloud MobileGateway resource: %s", err)
	}

	// connect to switch
	if switchID > 0 {
		_, err = client.MobileGateway.ConnectToSwitch(mgw.ID, sacloud.ID(switchID))
		if err != nil {
			return fmt.Errorf("Failed to create SakuraCloud MobileGateway resource: %s", err)
		}

		if len(mgw.Settings.MobileGateway.Interfaces) == 0 {
			mgw.Settings.MobileGateway.Interfaces = append(mgw.Settings.MobileGateway.Interfaces, nil)
		}
		mgw.SetPrivateInterface(ip, nwMaskLen)
		mgw, err = client.MobileGateway.Update(mgw.ID, mgw)
		if err != nil {
			return fmt.Errorf("MobileGatewayInterfaceConnect is failed: %s", err)
		}

		_, err = client.MobileGateway.Config(mgw.ID)
		if err != nil {
			return fmt.Errorf("MobileGatewayInterfaceConnect is failed: %s", err)
		}
	}

	rawTrafficControl := d.Get("traffic_control").([]interface{})
	if len(rawTrafficControl) > 0 {
		values := rawTrafficControl[0].(map[string]interface{})
		trafficControl := &sacloud.TrafficMonitoringConfig{
			TrafficQuotaInMB:     values["quota"].(int),
			BandWidthLimitInKbps: values["band_width_limit"].(int),
			EMailConfig: &sacloud.TrafficMonitoringNotifyEmail{
				Enabled: values["enable_email"].(bool),
			},
			SlackConfig: &sacloud.TrafficMonitoringNotifySlack{
				Enabled:             values["enable_slack"].(bool),
				IncomingWebhooksURL: values["slack_webhook"].(string),
			},
			AutoTrafficShaping: values["auto_traffic_shaping"].(bool),
		}

		if _, err := client.MobileGateway.SetTrafficMonitoringConfig(mgw.ID, trafficControl); err != nil {
			return fmt.Errorf("Failed to enable traffic-control on SakuraCloud MobileGateway: %s", err)
		}
	}

	// set DNS
	dns1 := d.Get("dns_server1").(string)
	dns2 := d.Get("dns_server2").(string)
	if dns1 != "" || dns2 != "" {
		_, err = client.MobileGateway.SetDNS(mgw.ID, sacloud.NewMobileGatewayResolver(dns1, dns2))
		if err != nil {
			return fmt.Errorf("Failed to wait SakuraCloud MobileGateway boot: %s", err)
		}
	}

	// static route
	if staticRoutes, ok := getListFromResource(d, "static_route"); ok && len(staticRoutes) > 0 {
		for _, rawStaticRoutes := range staticRoutes {
			values := mapToResourceData(rawStaticRoutes.(map[string]interface{}))
			staticRoute := expandMobileGatewayStaticRoute(values)

			// check duplicated
			for _, sr := range mgw.Settings.MobileGateway.StaticRoutes {
				if sr.Prefix == staticRoute.Prefix {
					return fmt.Errorf("prefix %q already exists", sr.Prefix)
				}
			}

			mgw.Settings.MobileGateway.StaticRoutes = append(mgw.Settings.MobileGateway.StaticRoutes, staticRoute)
		}
	}

	mgw, err = client.MobileGateway.UpdateSetting(mgw.ID, mgw)
	if err != nil {
		return fmt.Errorf("Failed to enable SakuraCloud MobileGatewayStaticRoute resource: %s", err)
	}
	_, err = client.MobileGateway.Config(mgw.ID)
	if err != nil {
		return fmt.Errorf("Couldn'd apply SakuraCloud MobileGateway config: %s", err)
	}

	// boot
	time.Sleep(90 * time.Second) // !HACK! For avoid that MobileGateway becomes an invalid state

	_, err = client.MobileGateway.Boot(mgw.ID)
	if err != nil {
		return fmt.Errorf("Failed to wait SakuraCloud MobileGateway boot: %s", err)
	}

	err = client.MobileGateway.SleepUntilUp(mgw.ID, client.DefaultTimeoutDuration)
	if err != nil {
		return fmt.Errorf("Failed to wait SakuraCloud MobileGateway boot: %s", err)
	}

	d.SetId(mgw.GetStrID())
	return resourceSakuraCloudMobileGatewayRead(d, meta)
}

func resourceSakuraCloudMobileGatewayRead(d *schema.ResourceData, meta interface{}) error {
	client := getSacloudAPIClient(d, meta)

	mgw, err := client.MobileGateway.Read(toSakuraCloudID(d.Id()))
	if err != nil {
		if sacloudErr, ok := err.(api.Error); ok && sacloudErr.ResponseCode() == 404 {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Couldn't find SakuraCloud MobileGateway resource: %s", err)
	}

	return setMobileGatewayResourceData(d, client, mgw)
}

func resourceSakuraCloudMobileGatewayUpdate(d *schema.ResourceData, meta interface{}) error {
	client := getSacloudAPIClient(d, meta)

	sakuraMutexKV.Lock(d.Id())
	defer sakuraMutexKV.Unlock(d.Id())

	mgw, err := client.MobileGateway.Read(toSakuraCloudID(d.Id()))
	if err != nil {
		return fmt.Errorf("Couldn't find SakuraCloud MobileGateway resource: %s", err)
	}

	var switchID sacloud.ID
	var ip string
	var nwMaskLen int
	if rawSwitchID, ok := d.GetOk("switch_id"); ok {
		strSwitchID := rawSwitchID.(string)
		if strSwitchID != "" {
			switchID = toSakuraCloudID(strSwitchID)
			if rawIP, ok := d.GetOk("private_ipaddress"); ok {
				ip = rawIP.(string)
			}
			if rawNWMaskLen, ok := d.GetOk("private_nw_mask_len"); ok {
				nwMaskLen = rawNWMaskLen.(int)
			}

			if ip == "" || nwMaskLen == 0 {
				return errors.New("MobileGateway needs private_ipaddress and private_nw_mask_len when switch_id is specified")
			}
		}
	}

	if d.HasChange("name") {
		mgw.Name = d.Get("name").(string)
	}
	if d.HasChange("icon_id") {
		if iconID, ok := d.GetOk("icon_id"); ok {
			mgw.SetIconByID(toSakuraCloudID(iconID.(string)))
		} else {
			mgw.ClearIcon()
		}
	}
	if d.HasChange("description") {
		if description, ok := d.GetOk("description"); ok {
			mgw.Description = description.(string)
		} else {
			mgw.Description = ""
		}
	}

	if d.HasChange("tags") {
		rawTags := d.Get("tags").([]interface{})
		if rawTags != nil {
			mgw.Tags = expandTags(client, rawTags)
		} else {
			mgw.Tags = expandTags(client, []interface{}{})
		}
	}

	if d.HasChange("internet_connection") {
		mgw.Settings.MobileGateway.InternetConnection.Enabled = "False"
		if d.Get("internet_connection").(bool) {
			mgw.Settings.MobileGateway.InternetConnection.Enabled = "True"
		}
	}

	mgw, err = client.MobileGateway.Update(mgw.ID, mgw)
	if err != nil {
		return fmt.Errorf("Error updating SakuraCloud MobileGateway resource: %s", err)
	}

	// need shutdown fields
	needRestart := false
	if d.HasChanges("switch_id", "private_ipaddress", "private_nw_mask_len") {
		// shutdown required for changing network settings
		if mgw.IsUp() {
			needRestart = true

			err = handleShutdown(client.MobileGateway, mgw.ID, d, client.DefaultTimeoutDuration)
			if err != nil {
				return fmt.Errorf("Error updating SakuraCloud MobileGateway resource: %s", err)
			}
			err = client.MobileGateway.SleepUntilDown(mgw.ID, client.DefaultTimeoutDuration)
			if err != nil {
				return fmt.Errorf("Error updating SakuraCloud MobileGateway resource: %s", err)
			}
		}

		// disconnect from switch if already connected
		if d.HasChange("switch_id") && len(mgw.Interfaces) > 1 && mgw.Interfaces[1].Switch != nil {
			_, err = client.MobileGateway.DisconnectFromSwitch(mgw.ID)
			if err != nil {
				return fmt.Errorf("Error updating SakuraCloud MobileGateway resource: %s", err)
			}
		}

		if switchID > 0 {
			_, err = client.MobileGateway.ConnectToSwitch(mgw.ID, switchID)
			if err != nil {
				return fmt.Errorf("Error updating SakuraCloud MobileGateway resource: %s", err)
			}

			if len(mgw.Settings.MobileGateway.Interfaces) == 0 {
				mgw.Settings.MobileGateway.Interfaces = append(mgw.Settings.MobileGateway.Interfaces, nil)
			}
			mgw.SetPrivateInterface(ip, nwMaskLen)
		} else {
			mgw.ClearPrivateInterface()
		}

		mgw, err = client.MobileGateway.Update(mgw.ID, mgw)
		if err != nil {
			return fmt.Errorf("MobileGatewayInterfaceConnect is failed: %s", err)
		}

		_, err = client.MobileGateway.Config(mgw.ID)
		if err != nil {
			return fmt.Errorf("MobileGatewayInterfaceConnect is failed: %s", err)
		}
	}

	if d.HasChange("traffic_control") {
		rawTrafficControl := d.Get("traffic_control").([]interface{})
		if len(rawTrafficControl) > 0 {
			values := rawTrafficControl[0].(map[string]interface{})
			trafficControl := &sacloud.TrafficMonitoringConfig{
				TrafficQuotaInMB:     values["quota"].(int),
				BandWidthLimitInKbps: values["band_width_limit"].(int),
				EMailConfig: &sacloud.TrafficMonitoringNotifyEmail{
					Enabled: values["enable_email"].(bool),
				},
				SlackConfig: &sacloud.TrafficMonitoringNotifySlack{
					Enabled:             values["enable_slack"].(bool),
					IncomingWebhooksURL: values["slack_webhook"].(string),
				},
				AutoTrafficShaping: values["auto_traffic_shaping"].(bool),
			}

			if _, err := client.MobileGateway.SetTrafficMonitoringConfig(mgw.ID, trafficControl); err != nil {
				return fmt.Errorf("Failed to enable traffic-control on SakuraCloud MobileGateway: %s", err)
			}
		} else {
			if _, err := client.MobileGateway.DisableTrafficMonitoringConfig(mgw.ID); err != nil {
				if e, ok := err.(api.Error); !ok || e.ResponseCode() != http.StatusNotFound {
					return fmt.Errorf("Failed to disable traffic-control on SakuraCloud MobileGateway: %s", err)
				}
			}
		}
	}

	if d.HasChanges("dns1", "dns2") {
		dns1 := d.Get("dns_server1").(string)
		dns2 := d.Get("dns_server2").(string)
		if dns1 != "" || dns2 != "" {
			_, err = client.MobileGateway.SetDNS(mgw.ID, sacloud.NewMobileGatewayResolver(dns1, dns2))
			if err != nil {
				return fmt.Errorf("Failed to wait SakuraCloud MobileGateway boot: %s", err)
			}
		}
	}

	// static route
	if d.HasChange("static_route") {
		mgw.Settings.MobileGateway.StaticRoutes = []*sacloud.MGWStaticRoute{}
		if staticRoutes, ok := getListFromResource(d, "static_route"); ok && len(staticRoutes) > 0 {
			for _, rawStaticRoutes := range staticRoutes {
				values := mapToResourceData(rawStaticRoutes.(map[string]interface{}))
				staticRoute := expandMobileGatewayStaticRoute(values)

				// check duplicated
				for _, sr := range mgw.Settings.MobileGateway.StaticRoutes {
					if sr.Prefix == staticRoute.Prefix {
						return fmt.Errorf("prefix %q already exists", sr.Prefix)
					}
				}
				mgw.Settings.MobileGateway.StaticRoutes = append(mgw.Settings.MobileGateway.StaticRoutes, staticRoute)
			}
		}

		mgw, err = client.MobileGateway.UpdateSetting(mgw.ID, mgw)
		if err != nil {
			return fmt.Errorf("Failed to enable SakuraCloud MobileGatewayStaticRoute resource: %s", err)
		}
		_, err = client.MobileGateway.Config(mgw.ID)
		if err != nil {
			return fmt.Errorf("Couldn'd apply SakuraCloud MobileGateway config: %s", err)
		}
	}

	if needRestart {
		_, err = client.MobileGateway.Boot(mgw.ID)
		if err != nil {
			return fmt.Errorf("Error updating SakuraCloud MobileGateway resource: %s", err)
		}
		err = client.MobileGateway.SleepUntilUp(mgw.ID, client.DefaultTimeoutDuration)
		if err != nil {
			return fmt.Errorf("Error updating SakuraCloud MobileGateway resource: %s", err)
		}
	}

	return resourceSakuraCloudMobileGatewayRead(d, meta)
}

func resourceSakuraCloudMobileGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	client := getSacloudAPIClient(d, meta)

	sakuraMutexKV.Lock(d.Id())
	defer sakuraMutexKV.Unlock(d.Id())

	mgw, err := client.MobileGateway.Read(toSakuraCloudID(d.Id()))
	if err != nil {
		if sacloudErr, ok := err.(api.Error); ok && sacloudErr.ResponseCode() == 404 {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Couldn't find SakuraCloud MobileGateway resource: %s", err)
	}

	err = handleShutdown(client.MobileGateway, toSakuraCloudID(d.Id()), d, client.DefaultTimeoutDuration)
	if err != nil {
		return fmt.Errorf("Error stopping SakuraCloud MobileGateway resource: %s", err)
	}

	// delete SIMs
	sims, err := client.MobileGateway.ListSIM(mgw.ID, nil)
	if err != nil {
		return fmt.Errorf("Error deleting SakuraCloud MobileGateway resource: %s", err)
	}

	for _, sim := range sims {
		if sim.Activated {
			_, err = client.SIM.Deactivate(toSakuraCloudID(sim.ResourceID))
			if err != nil {
				return fmt.Errorf("Failed to deactivate SakuraCloud SIM resource: %s", err)
			}
		}
		_, err = client.SIM.ClearIP(toSakuraCloudID(sim.ResourceID))
		if err != nil {
			return fmt.Errorf("Failed to ClearIP SakuraCloud SIM resource: %s", err)
		}
		_, err = client.MobileGateway.DeleteSIM(mgw.ID, toSakuraCloudID(sim.ResourceID))
		if err != nil {
			return fmt.Errorf("Error deleting SakuraCloud MobileGateway resource: %s", err)
		}
	}

	_, err = client.MobileGateway.Delete(toSakuraCloudID(d.Id()))
	if err != nil {
		return fmt.Errorf("Error deleting SakuraCloud MobileGateway resource: %s", err)
	}

	return nil
}

func setMobileGatewayResourceData(d *schema.ResourceData, client *APIClient, data *sacloud.MobileGateway) error {
	if data.IsFailed() {
		d.SetId("")
		return fmt.Errorf("MobileGateway[%d] state is failed", data.ID)
	}

	d.Set("public_ipaddress", data.Interfaces[0].IPAddress)
	d.Set("public_nw_mask_len", data.Interfaces[0].Switch.Subnet.NetworkMaskLen)
	d.Set("internet_connection", strings.ToLower(data.Settings.MobileGateway.InternetConnection.Enabled) == "true")

	if len(data.Interfaces) > 1 && data.Interfaces[1].Switch != nil {
		d.Set("switch_id", data.Interfaces[1].Switch.GetStrID())
		d.Set("private_ipaddress", data.Settings.MobileGateway.Interfaces[1].IPAddress[0])
		d.Set("private_nw_mask_len", data.Settings.MobileGateway.Interfaces[1].NetworkMaskLen)
	} else {
		d.Set("switch_id", "")
		d.Set("private_ipaddress", "")
		d.Set("private_nw_mask_len", "")
	}

	tc, err := client.MobileGateway.GetTrafficMonitoringConfig(data.ID)
	if err != nil {
		if e, ok := err.(api.Error); ok && e.ResponseCode() != http.StatusNotFound {
			return fmt.Errorf("Error reading SakuraCloud MobileGateway resource(traffic-control): %s", err)
		}
	}

	if tc != nil {
		tcValues := map[string]interface{}{
			"quota":                tc.TrafficQuotaInMB,
			"band_width_limit":     tc.BandWidthLimitInKbps,
			"auto_traffic_shaping": tc.AutoTrafficShaping,
		}
		if tc.EMailConfig == nil {
			tcValues["enable_email"] = false
		} else {
			tcValues["enable_email"] = tc.EMailConfig.Enabled
		}
		if tc.SlackConfig == nil {
			tcValues["enable_slack"] = false
			tcValues["slack_webhook"] = ""
		} else {
			tcValues["enable_slack"] = tc.SlackConfig.Enabled
			tcValues["slack_webhook"] = tc.SlackConfig.IncomingWebhooksURL
		}
		d.Set("traffic_control", []interface{}{tcValues})
	}

	resolver, err := client.MobileGateway.GetDNS(data.ID)
	if err != nil {
		return fmt.Errorf("Error reading SakuraCloud MobileGateway resource(dns-resolver): %s", err)
	}

	d.Set("dns_server1", resolver.SimGroup.DNS1)
	d.Set("dns_server2", resolver.SimGroup.DNS2)

	var staticRoutes []map[string]interface{}
	if data.HasStaticRoutes() {
		for _, r := range data.Settings.MobileGateway.StaticRoutes {
			staticRoutes = append(staticRoutes, map[string]interface{}{
				"prefix":   r.Prefix,
				"next_hop": r.NextHop,
			})
		}
	}
	d.Set("static_route", staticRoutes)

	d.Set("name", data.Name)
	d.Set("icon_id", data.GetIconStrID())
	d.Set("description", data.Description)
	d.Set("tags", data.Tags)

	sims, err := client.MobileGateway.ListSIM(data.ID, nil)
	if err != nil {
		return fmt.Errorf("Error reading SakuraCloud MobileGateway resource(dns-resolver): %s", err)
	}
	simIDs := []string{}
	for _, sim := range sims {
		simIDs = append(simIDs, sim.ResourceID)
	}
	d.Set("sim_ids", simIDs)

	setPowerManageTimeoutValueToState(d)

	return nil
}

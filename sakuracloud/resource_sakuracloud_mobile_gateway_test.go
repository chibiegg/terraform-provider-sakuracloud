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
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/sacloud/libsacloud/v2/sacloud"
)

func TestAccSakuraCloudMobileGateway_basic(t *testing.T) {
	resourceName := "sakuracloud_mobile_gateway.foobar"
	name := randomName()

	var mgw sacloud.MobileGateway
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckSakuraCloudIconDestroy,
			testCheckSakuraCloudMobileGatewayDestroy,
			testCheckSakuraCloudSwitchDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: buildConfigWithArgs(testAccSakuraCloudMobileGateway_basic, name),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraCloudMobileGatewayExists(resourceName, &mgw),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "internet_connection", "true"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.4151227546", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1852302624", "tag2"),
					resource.TestCheckResourceAttrPair(
						resourceName, "private_network_interface.0.switch_id",
						"sakuracloud_switch.foobar", "id",
					),
					resource.TestCheckResourceAttr(resourceName, "private_network_interface.0.ip_address", "192.168.11.101"),
					resource.TestCheckResourceAttr(resourceName, "private_network_interface.0.netmask", "24"),
					resource.TestCheckResourceAttrPair(
						resourceName, "icon_id",
						"sakuracloud_icon.foobar", "id",
					),
					resource.TestCheckResourceAttr(resourceName, "traffic_control.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "traffic_control.0.quota", "256"),
					resource.TestCheckResourceAttr(resourceName, "traffic_control.0.band_width_limit", "64"),
					resource.TestCheckResourceAttr(resourceName, "traffic_control.0.enable_email", "true"),
					resource.TestCheckResourceAttr(resourceName, "traffic_control.0.enable_slack", "true"),
					resource.TestCheckResourceAttr(resourceName, "traffic_control.0.slack_webhook", "https://hooks.slack.com/services/xxx/xxx/xxx"),
					resource.TestCheckResourceAttr(resourceName, "traffic_control.0.auto_traffic_shaping", "true"),
					resource.TestCheckResourceAttr(resourceName, "static_route.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "static_route.0.prefix", "192.168.10.0/24"),
					resource.TestCheckResourceAttr(resourceName, "static_route.0.next_hop", "192.168.11.1"),
					resource.TestCheckResourceAttr(resourceName, "static_route.1.prefix", "192.168.10.0/25"),
					resource.TestCheckResourceAttr(resourceName, "static_route.1.next_hop", "192.168.11.2"),
					resource.TestCheckResourceAttr(resourceName, "static_route.2.prefix", "192.168.10.0/26"),
					resource.TestCheckResourceAttr(resourceName, "static_route.2.next_hop", "192.168.11.3"),
				),
			},
			{
				Config: buildConfigWithArgs(testAccSakuraCloudMobileGateway_update, name),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraCloudMobileGatewayExists(resourceName, &mgw),
					resource.TestCheckResourceAttr(resourceName, "name", name+"-upd"),
					resource.TestCheckResourceAttr(resourceName, "description", "description-upd"),
					resource.TestCheckResourceAttr(resourceName, "internet_connection", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.2362157161", "tag1-upd"),
					resource.TestCheckResourceAttr(resourceName, "tags.3412841145", "tag2-upd"),
					resource.TestCheckResourceAttr(resourceName, "icon_id", ""),
					resource.TestCheckResourceAttr(resourceName, "traffic_control.#", "0"),
					resource.TestCheckResourceAttrPair(
						resourceName, "private_network_interface.0.switch_id",
						"sakuracloud_switch.foobar", "id"),
					resource.TestCheckResourceAttr(resourceName, "private_network_interface.0.ip_address", "192.168.11.101"),
					resource.TestCheckResourceAttr(resourceName, "private_network_interface.0.netmask", "24"),
					resource.TestCheckResourceAttr(resourceName, "static_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sim.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sim_route.#", "0"),
				),
			},
		},
	})
}

func testCheckSakuraCloudMobileGatewayExists(n string, mgs *sacloud.MobileGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no MobileGateway ID is set")
		}

		client := testAccProvider.Meta().(*APIClient)
		zone := rs.Primary.Attributes["zone"]
		mgwOp := sacloud.NewMobileGatewayOp(client)

		foundMobileGateway, err := mgwOp.Read(context.Background(), zone, sakuraCloudID(rs.Primary.ID))

		if err != nil {
			return err
		}

		if foundMobileGateway.ID.String() != rs.Primary.ID {
			return fmt.Errorf("not found MobileGateway: %s", rs.Primary.ID)
		}

		*mgs = *foundMobileGateway

		return nil
	}
}

func testCheckSakuraCloudMobileGatewayDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*APIClient)
	mgwOp := sacloud.NewMobileGatewayOp(client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakuracloud_mobile_gateway" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		zone := rs.Primary.Attributes["zone"]
		_, err := mgwOp.Read(context.Background(), zone, sakuraCloudID(rs.Primary.ID))

		if err == nil {
			return fmt.Errorf("still exists MobileGateway: %s", rs.Primary.ID)
		}
	}

	return nil
}

func TestAccImportSakuraCloudMobileGateway_basic(t *testing.T) {
	rand := randomName()

	checkFn := func(s []*terraform.InstanceState) error {
		if len(s) != 1 {
			return fmt.Errorf("expected 1 state: %#v", s)
		}
		expects := map[string]string{
			"name":                                   rand,
			"private_network_interface.0.ip_address": "192.168.11.101",
			"private_network_interface.0.netmask":    "24",
			"description":                            "description",
			"internet_connection":                    "true",
			"inter_device_communication":             "false",
			"tags.4151227546":                        "tag1",
			"tags.1852302624":                        "tag2",
			"traffic_control.0.quota":                "256",
			"traffic_control.0.band_width_limit":     "64",
			"traffic_control.0.enable_email":         "true",
			"traffic_control.0.enable_slack":         "true",
			"traffic_control.0.slack_webhook":        "https://hooks.slack.com/services/xxx/xxx/xxx",
			"traffic_control.0.auto_traffic_shaping": "true",
			"static_route.0.prefix":                  "192.168.10.0/24",
			"static_route.0.next_hop":                "192.168.11.1",
			"static_route.1.prefix":                  "192.168.10.0/25",
			"static_route.1.next_hop":                "192.168.11.2",
			"static_route.2.prefix":                  "192.168.10.0/26",
			"static_route.2.next_hop":                "192.168.11.3",
		}

		if err := compareStateMulti(s[0], expects); err != nil {
			return err
		}
		return stateNotEmptyMulti(s[0], "private_network_interface.0.switch_id", "icon_id")
	}

	resourceName := "sakuracloud_mobile_gateway.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckSakuraCloudMobileGatewayDestroy,
			testCheckSakuraCloudSwitchDestroy,
			testCheckSakuraCloudIconDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: buildConfigWithArgs(testAccSakuraCloudMobileGateway_basic, rand),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateCheck:  checkFn,
				ImportStateVerify: true,
			},
		},
	})
}

const testAccSakuraCloudMobileGateway_basic = `
data sakuracloud_zone "zone" {}

resource "sakuracloud_switch" "foobar" {
  name = "{{ .arg0 }}"
}

resource "sakuracloud_mobile_gateway" "foobar" {
  private_network_interface {
    switch_id  = sakuracloud_switch.foobar.id
    ip_address = "192.168.11.101"
    netmask    = 24
  }
  internet_connection = true
  name                = "{{ .arg0 }}"
  description         = "description"
  tags                = ["tag1", "tag2"]
  icon_id             = sakuracloud_icon.foobar.id
  dns_servers         = data.sakuracloud_zone.zone.dns_servers

  traffic_control {
    quota                = 256
    band_width_limit     = 64
    enable_email         = true
    enable_slack         = true
    slack_webhook        = "https://hooks.slack.com/services/xxx/xxx/xxx"
    auto_traffic_shaping = true
  }

  static_route {
    prefix   = "192.168.10.0/24"
    next_hop = "192.168.11.1"
  }
  static_route {
    prefix   = "192.168.10.0/25"
    next_hop = "192.168.11.2"
  }
  static_route {
    prefix   = "192.168.10.0/26"
    next_hop = "192.168.11.3"
  }
}

resource "sakuracloud_icon" "foobar" {
  name          = "{{ .arg0 }}"
  base64content = "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"
}
`

const testAccSakuraCloudMobileGateway_update = `
data sakuracloud_zone "zone" {}

resource "sakuracloud_switch" "foobar" {
  name = "{{ .arg0 }}"
}
resource "sakuracloud_mobile_gateway" "foobar" {
  private_network_interface {
    switch_id  = sakuracloud_switch.foobar.id
    ip_address = "192.168.11.101"
    netmask    = 24
  }
  internet_connection = false
  name                = "{{ .arg0 }}-upd"
  description         = "description-upd"
  tags                = ["tag1-upd", "tag2-upd"]
  dns_servers         = data.sakuracloud_zone.zone.dns_servers
}`

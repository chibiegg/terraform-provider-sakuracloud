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
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/sacloud/libsacloud/sacloud"
)

func TestAccResourceSakuraCloudPacketFilterRule_basic(t *testing.T) {
	var filter sacloud.PacketFilter
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSakuraCloudPacketFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSakuraCloudPacketFilterRuleConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSakuraCloudPacketFilterExists("sakuracloud_packet_filter.foobar", &filter),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule0", "protocol", "tcp"),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule0", "source_nw", "192.168.2.0"),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule0", "source_port", "80"),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule0", "dest_port", "80"),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule0", "allow", "true"),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule0", "order", "0"),

					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule1", "protocol", "tcp"),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule1", "source_nw", "192.168.2.0"),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule1", "source_port", "443"),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule1", "dest_port", "443"),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule1", "allow", "true"),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule1", "order", "1"),

					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule2", "protocol", "ip"),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule2", "source_nw", ""),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule2", "source_port", ""),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule2", "dest_port", ""),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule2", "allow", "false"),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule2", "order", "2"),
				),
			},
			{
				Config: testAccCheckSakuraCloudPacketFilterRuleConfig_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule0", "protocol", "udp"),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule0", "source_nw", "192.168.2.2"),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule0", "source_port", "80"),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule0", "dest_port", "80"),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule0", "allow", "true"),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule0", "order", "0"),

					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule1", "protocol", "udp"),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule1", "source_nw", "192.168.2.2"),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule1", "source_port", "443"),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule1", "dest_port", "443"),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule1", "allow", "true"),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule1", "order", "1"),

					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule2", "protocol", "ip"),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule2", "source_nw", ""),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule2", "source_port", ""),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule2", "dest_port", ""),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule2", "allow", "false"),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule2", "order", "2"),
				),
			},
		},
	})
}

func TestAccResourceSakuraCloudPacketFilterRule_discontinuousIndex(t *testing.T) {
	var filter sacloud.PacketFilter
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSakuraCloudPacketFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSakuraCloudPacketFilterRuleConfig_discontinuous_index,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSakuraCloudPacketFilterExists("sakuracloud_packet_filter.foobar", &filter),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule0", "protocol", "tcp"),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule0", "source_nw", "192.168.2.0"),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule0", "source_port", "80"),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule0", "dest_port", "80"),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule0", "allow", "true"),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule0", "order", "0"),

					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule2", "protocol", "ip"),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule2", "source_nw", ""),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule2", "source_port", ""),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule2", "dest_port", ""),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule2", "allow", "false"),
					resource.TestCheckResourceAttr(
						"sakuracloud_packet_filter_rule.rule2", "order", "2"),
				),
			},
		},
	})
}

var testAccCheckSakuraCloudPacketFilterRuleConfig_basic = `
resource "sakuracloud_packet_filter" "foobar" {
    name = "mypacket_filter"
    description = "PacketFilter from TerraForm for SAKURA CLOUD"
}

resource sakuracloud_packet_filter_rule "rule0" {
    packet_filter_id = "${sakuracloud_packet_filter.foobar.id}"

 	protocol    = "tcp"
	source_nw   = "192.168.2.0"
	source_port = "80"
	dest_port   = "80"
	allow       = true
	order       = 0
}

resource sakuracloud_packet_filter_rule "rule1" {
    packet_filter_id = "${sakuracloud_packet_filter.foobar.id}"

	protocol    = "tcp"
	source_nw   = "192.168.2.0"
	source_port = "443"
	dest_port   = "443"
	allow       = true
	order       = 1
}

resource sakuracloud_packet_filter_rule "rule2" {
    packet_filter_id = "${sakuracloud_packet_filter.foobar.id}"

 	protocol    = "ip"
	allow       = false
	order       = 2
}
`

var testAccCheckSakuraCloudPacketFilterRuleConfig_update = `
resource "sakuracloud_packet_filter" "foobar" {
    name = "mypacket_filter"
    description = "PacketFilter from TerraForm for SAKURA CLOUD"
}

resource sakuracloud_packet_filter_rule "rule0" {
    packet_filter_id = "${sakuracloud_packet_filter.foobar.id}"

   	protocol    = "udp"
  	source_nw   = "192.168.2.2"
  	source_port = "80"
  	dest_port   = "80"
   	allow       = true
  	order       = 0
}

resource sakuracloud_packet_filter_rule "rule1" {
    packet_filter_id = "${sakuracloud_packet_filter.foobar.id}"

   	protocol    = "udp"
  	source_nw   = "192.168.2.2"
  	source_port = "443"
  	dest_port   = "443"
  	allow       = true
  	order       = 1
}

resource sakuracloud_packet_filter_rule "rule2" {
    packet_filter_id = "${sakuracloud_packet_filter.foobar.id}"

  	protocol    = "ip"
	allow       = false
 	order       = 2
}
`

var testAccCheckSakuraCloudPacketFilterRuleConfig_discontinuous_index = `
resource "sakuracloud_packet_filter" "foobar" {
    name = "mypacket_filter"
    description = "PacketFilter from TerraForm for SAKURA CLOUD"
}

resource sakuracloud_packet_filter_rule "rule0" {
    packet_filter_id = "${sakuracloud_packet_filter.foobar.id}"

 	protocol    = "tcp"
	source_nw   = "192.168.2.0"
	source_port = "80"
	dest_port   = "80"
	allow       = true
	order       = 0
}

resource sakuracloud_packet_filter_rule "rule2" {
    packet_filter_id = "${sakuracloud_packet_filter.foobar.id}"

 	protocol    = "ip"
	allow       = false
	order       = 2
}
`

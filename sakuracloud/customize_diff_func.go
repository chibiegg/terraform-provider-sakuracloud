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
	"reflect"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func hasTagResourceCustomizeDiff(d *schema.ResourceDiff, meta interface{}) error {
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if o != nil && n != nil {
			os := expandStringList(o.([]interface{}))
			ns := expandStringList(n.([]interface{}))

			sort.Strings(os)
			sort.Strings(ns)
			if reflect.DeepEqual(os, ns) {
				return d.Clear("tags")
			}
		}
	}
	return nil
}

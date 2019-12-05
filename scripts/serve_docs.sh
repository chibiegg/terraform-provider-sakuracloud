#!/bin/bash
# Copyright 2016-2019 terraform-provider-sakuracloud authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


set -e

DOCKER_IMAGE_NAME="sacloud/mkdocs:latest"
DOCKER_CONTAINER_NAME="terraform-provider-sakuracloud-docs-container"

if [[ $(docker ps -a | grep $DOCKER_CONTAINER_NAME) != "" ]]; then
  docker rm -f $DOCKER_CONTAINER_NAME 2>/dev/null
fi


docker run --name $DOCKER_CONTAINER_NAME \
       -v $PWD/build_docs:/workdir \
       -p 80:80 \
       $DOCKER_IMAGE_NAME serve --dev-addr=0.0.0.0:80

docker rm -f $DOCKER_CONTAINER_NAME 2>/dev/null

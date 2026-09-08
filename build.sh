#!/bin/sh

set -eu

image_name="${IMAGE_NAME:-ascii-art-web-docker}"
container_name="${CONTAINER_NAME:-dockerize}"
host_port="${HOST_PORT:-8080}"

docker image build -f Dockerfile -t "$image_name" .

if docker container inspect "$container_name" >/dev/null 2>&1; then
    docker container rm --force "$container_name" >/dev/null
fi

docker container run \
    --detach \
    --name "$container_name" \
    --publish "$host_port:8080" \
    "$image_name"

printf 'Container %s is running at http://localhost:%s\n' "$container_name" "$host_port"

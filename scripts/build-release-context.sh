#!/usr/bin/env bash
set -euo pipefail

repository_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
release_version=${1:-$(tr -d '[:space:]' < "$repository_root/RELEASE_VERSION")}
target_arch=${TARGETARCH:-$(go env GOARCH)}
release_context="$repository_root/dist/release-context"

if [[ ! "$release_version" =~ ^[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z]+)*$ ]]; then
  echo "Invalid release version: $release_version" >&2
  exit 1
fi

cd "$repository_root"

echo "Building Portainer $release_version for linux/$target_arch"
CI=true pnpm install --frozen-lockfile
NODE_ENV=production pnpm run build --config webpack/webpack.production.js
CONTAINER_IMAGE_TAG="$release_version" SKIP_GO_GET=true ./build/build_binary.sh linux "$target_arch"

rm -rf "$release_context"
mkdir -p "$release_context"
cp dist/portainer "$release_context/portainer"
cp -R dist/public "$release_context/public"
cp -R dist/mustache-templates "$release_context/mustache-templates"
cp build/release-context.Dockerfile "$release_context/Dockerfile"

git_commit=$(git rev-parse --short HEAD)
build_date=$(date -u +%Y-%m-%dT%H:%M:%SZ)

cat > "$release_context/BUILD-IMAGE.txt" <<EOF
docker build --platform linux/$target_arch --build-arg GIT_COMMIT=$git_commit --build-arg BUILD_DATE=$build_date -t themodcrafttmc/portainer:$release_version -t themodcrafttmc/portainer:latest .
docker push themodcrafttmc/portainer:$release_version
docker push themodcrafttmc/portainer:latest
EOF

echo
echo "Release context ready: $release_context"
echo "Run:"
echo "  cd $release_context"
cat "$release_context/BUILD-IMAGE.txt"

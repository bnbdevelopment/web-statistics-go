#!/usr/bin/env bash
set -euo pipefail

base_ref="${1:?usage: bump-versions.sh <base-ref> <head-ref> [output-file]}"
head_ref="${2:?usage: bump-versions.sh <base-ref> <head-ref> [output-file]}"
output_file="${3:-/dev/stdout}"

if ! git cat-file -e "${base_ref}^{commit}" 2>/dev/null; then
  base_ref="$(git rev-list --max-parents=0 "${head_ref}")"
fi

mapfile -t changed_files < <(git diff --name-only "${base_ref}" "${head_ref}")
backend=false
frontend=false
chart=false

for file in "${changed_files[@]}"; do
  case "$file" in
    backend/*) backend=true ;;
    frontend/*) frontend=true ;;
    charts/web-statistics/*) chart=true ;;
  esac
done

# The chart consumes the application images, so a new image version also
# produces a chart version whose defaults reference that immutable image tag.
if [[ "$backend" == true || "$frontend" == true ]]; then
  chart=true
fi

if [[ "$backend" == false && "$frontend" == false && "$chart" == false ]]; then
  {
    echo "changed=false"
    echo 'images=[]'
    echo "chart=false"
    echo 'releases=[]'
  } >> "$output_file"
  exit 0
fi

commit_text="$(git log --format='%s%n%b' "${base_ref}..${head_ref}")"
level='patch'
if grep -Eq '(^|[[:space:]])BREAKING[ -]CHANGE:|^[a-zA-Z]+(\([^)]*\))?!:' <<< "$commit_text"; then
  level='major'
elif grep -Eq '^feat(\([^)]*\))?:' <<< "$commit_text"; then
  level='minor'
fi

bump() {
  local current="$1"
  local major minor patch
  IFS=. read -r major minor patch <<< "$current"
  [[ "$major" =~ ^[0-9]+$ && "$minor" =~ ^[0-9]+$ && "$patch" =~ ^[0-9]+$ ]] || {
    echo "invalid semantic version: $current" >&2
    exit 1
  }
  case "$level" in
    major) printf '%s.0.0' "$((major + 1))" ;;
    minor) printf '%s.%s.0' "$major" "$((minor + 1))" ;;
    patch) printf '%s.%s.%s' "$major" "$minor" "$((patch + 1))" ;;
  esac
}

releases=()
images=()

if [[ "$backend" == true ]]; then
  backend_version="$(bump "$(tr -d '[:space:]' < backend/VERSION)")"
  printf '%s\n' "$backend_version" > backend/VERSION
  images+=(backend)
  releases+=("backend:$backend_version")
fi

if [[ "$frontend" == true ]]; then
  current_frontend="$(node -p "require('./frontend/package.json').version")"
  frontend_version="$(bump "$current_frontend")"
  npm version "$frontend_version" --no-git-tag-version --prefix frontend >/dev/null
  images+=(frontend)
  releases+=("frontend:$frontend_version")
fi

if [[ "$chart" == true ]]; then
  current_chart="$(sed -n 's/^version:[[:space:]]*//p' charts/web-statistics/Chart.yaml)"
  chart_version="$(bump "$current_chart")"
  sed -i "s/^version:.*/version: ${chart_version}/" charts/web-statistics/Chart.yaml
  sed -i "s/^appVersion:.*/appVersion: \"${chart_version}\"/" charts/web-statistics/Chart.yaml
  releases+=("chart:$chart_version")
fi

if [[ "$backend" == true ]]; then
  sed -i "s/^    tag:.*# backend-version$/    tag: \"${backend_version}\" # backend-version/" charts/web-statistics/values.yaml
fi
if [[ "$frontend" == true ]]; then
  sed -i "s/^    tag:.*# frontend-version$/    tag: \"${frontend_version}\" # frontend-version/" charts/web-statistics/values.yaml
fi

json_array() {
  local first=true item
  printf '['
  for item in "$@"; do
    if [[ "$first" == false ]]; then printf ','; fi
    printf '"%s"' "$item"
    first=false
  done
  printf ']'
}

release_json='['
first=true
for release in "${releases[@]}"; do
  component="${release%%:*}"
  version="${release#*:}"
  [[ "$first" == true ]] || release_json+=','
  release_json+="{\"component\":\"${component}\",\"version\":\"${version}\",\"tag\":\"${component}-v${version}\"}"
  first=false
done
release_json+=']'

{
  echo "changed=true"
  echo "images=$(json_array "${images[@]}")"
  echo "chart=$chart"
  echo "releases=$release_json"
  echo "bump_level=$level"
} >> "$output_file"

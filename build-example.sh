#!/bin/sh

get_argument() {
  name="$1"
  default="$2"
  shift 2

  while [[ $# -gt 0 ]]; do
    if [[ "$1" == "$name" ]]; then value="$2"; break; fi
    if [[ "$1" == "${name}="* ]]; then value="${1#*=}"; break; fi
    shift
  done

  echo ${value:-"$default"}
}

version=$(get_argument "--version" "1.0.0" $@)
build_number=$(date +%Y%m%d%H%M)
branch=$(get_argument "--branch" "master" $@)

post_slack "\`\`\`
Start build:
  Version: $version ($build_number)
  Branch: $branch
\`\`\`"

sleep 20

post_slack "\`\`\`
Finished build:
  Version: $version ($build_number)
  Branch: $branch
\`\`\`"

# error
#build-ios

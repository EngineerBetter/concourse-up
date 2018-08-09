#!/bin/bash

set -ex
[ "$VERBOSE" ] && { set -x; export BOSH_LOG_LEVEL=debug; export BOSH_LOG_PATH=bosh.log; }

cp binary-linux/concourse-up-linux-amd64 ./cup
chmod +x ./cup

delete_bucket() {
  # attempt to delete the bucket
  # eventual consistency means it may already be gone
  # so continue regardless of result
  aws rb --force "s3://$1" 2>/dev/null || true
}

aws s3 ls \
| grep -E 'concourse-up-systest' \
| awk '{print $3}' \
> buckets

while read -r line; do
  if aws s3 ls "s3://$line" 2>/dev/null; then
    if [[ "$(aws s3 ls "s3://$line")" =~ terraform.tfstate ]]; then
      echo "$line" >> non-empty
    else
      delete_bucket "$line"
    fi
  else
    delete_bucket "$line"
  fi
done < buckets

while read -r line; do
  if echo "$line" | grep -qE '^concourse-up-systest-[0-9]+'; then
    echo "$line" | awk -F- '{print "yes yes | ./cup destroy --region "$5"-"$6"-"$7" "$3"-"$4}' >> cup-delete
  elif echo "$line" | grep -qE '^concourse-up-systest-[a-zA-Z]+-[0-9]+'; then
    echo "$line" | awk -F- '{print "yes yes | ./cup destroy --region "$6"-"$7"-"$8" "$3"-"$4"-"$5}' >> cup-delete
  else
    printf "Unexpected bucket format %s -- skipping\\n" "$line"
  fi
done < non-empty

sort -u cup-delete \
| xargs -P 8 -I {} bash -c '{}'

rm buckets non-empty cup-delete

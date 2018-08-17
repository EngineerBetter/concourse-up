#!/usr/bin/env ruby

`cp binary-linux/concourse-up-linux-amd64 ./cup`
`chmod +x ./cup`

non_empty = []

buckets = `aws s3 ls | grep -E 'concourse-up-systest' | awk '{print $3}'`

buckets.each_line do |bucket|
  bucket = bucket.strip
  deployment = bucket.split('-')[0..4].join('-')
  region = bucket.split('-')[5..7].join('-')

  `aws s3 ls s3://#{bucket} > /dev/null`
  if $? == 0
    contents = `aws s3 ls s3://#{bucket}`
    if contents.include?('terraform.tfstate') && contents.include?('director-state.json')
      non_empty.push(bucket)
    else
      puts "#{bucket} is missing key files"
      puts "Attempting to delete VPC"
      puts "==================================================================="
      puts "MANUAL CLEANUP MAY BE REQUIRED FOR #{bucket}"
      puts "==================================================================="
      vpc_id = `aws ec2 describe-vpcs --filters 'Name=tag:Name,Values=#{deployment}' --region #{region} | jq -r '.Vpcs[0].VpcId'`
      if vpc_id
        `aws --region #{region} ec2 delete-vpc --vpc-id #{vpc_id}`
      else
        puts "vpc for #{deployment} not found"
      end
    end
  else
    printf "#{bucket} doesn't really exist\n"
  end
end

return if non_empty.empty?

non_empty.each do |bucket|
  if /^concourse-up-systest-[0-9]+/ =~ "#{bucket}"
    deployment = bucket.split('-')[2..3].join('-')
    region = bucket.split('-')[4..6].join('-')
    str = sprintf("./cup destroy --region %s %s", region, deployment)
    `echo #{str} >> to_delete`
  elsif /^concourse-up-systest-[a-zA-Z]+-[0-9]+/ =~ "#{bucket}"
    deployment = bucket.split('-')[2..4].join('-')
    region = bucket.split('-')[5..7].join('-')
    str = sprintf("./cup destroy --region %s %s", region, deployment)
    `echo #{str} >> to_delete`
  else
    puts "Unexpected bucket format #{bucket} -- skipping"
  end
end

`sort -u to_delete \
 | xargs -P 8 -I {} bash -c '{}'`

`rm -f to_delete`

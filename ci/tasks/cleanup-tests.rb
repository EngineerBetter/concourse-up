#!/usr/bin/env ruby

`cp binary-linux/concourse-up-linux-amd64 ./cup`
`chmod +x ./cup`

non_empty = []

buckets = `aws s3 ls | grep -E 'concourse-up-systest' | awk '{print $3}'`

if buckets.empty?
  puts 'No residual systest buckets found'
  exit(0)
end

buckets.each_line do |bucket|
  bucket = bucket.strip
  deployment = bucket.split('-')[0..4].join('-')
  region = bucket.split('-')[5..7].join('-')

  aws_region_cmd = "aws --region #{region}"

  `aws s3 ls s3://#{bucket} > /dev/null`
  if $?.success?
    contents = `aws s3 ls s3://#{bucket}`
    if contents.include?('terraform.tfstate') && contents.include?('director-state.json')
      non_empty.push(bucket)
    else
      puts "#{bucket} is missing key files"
      puts 'Attempting to delete VPC'
      puts '==================================================================='
      puts "MANUAL CLEANUP MAY BE REQUIRED FOR #{bucket}"
      puts '==================================================================='
      vpc_id = `#{aws_region_cmd} ec2 describe-vpcs --filters 'Name=tag:Name,Values=#{deployment}' | jq -r '.Vpcs[0].VpcId'`
      if vpc_id
        rds_instance = `#{aws_region_cmd} rds describe-db-instances | jq -r '.DBInstances[] | select(.DBSubnetGroup.VpcId == "#{vpc_id}") | .DBInstanceIdentifier'`
        if rds_instance
          `#{aws_region_cmd} rds delete-db-instance --db-instance-identifier "#{rds_instance}" --skip-final-snapshot true`
          `#{aws_region_cmd} rds wait db-instance-deleted --db-instance-identifier "#{rds_instance}"`
        end
        `#{aws_region_cmd} ec2 delete-vpc --vpc-id #{vpc_id}`
      else
        puts "vpc for #{deployment} not found"
      end
      `aws s3 rb s3://#{bucket} --force`
    end
  else
    printf "#{bucket} doesn't really exist\n"
  end
end

exit(0) if non_empty.empty?

non_empty.each do |bucket|
  if /^concourse-up-systest-[0-9]+/ =~ "#{bucket}"
    deployment = bucket.split('-')[2..3].join('-')
    region = bucket.split('-')[4..6].join('-')
    str = sprintf("yes yes | ./cup destroy --region %s %s", region, deployment)
    `echo #{str} >> to_delete`
  elsif /^concourse-up-systest-[a-zA-Z]+-[0-9]+/ =~ "#{bucket}"
    deployment = bucket.split('-')[2..4].join('-')
    region = bucket.split('-')[5..7].join('-')
    str = sprintf("yes yes | ./cup destroy --region %s %s", region, deployment)
    `echo #{str} >> to_delete`
  else
    puts "Unexpected bucket format #{bucket} -- skipping"
  end
end

`sort -u to_delete \
 | xargs -P 8 -I {} bash -c '{}'`

`rm -f to_delete`

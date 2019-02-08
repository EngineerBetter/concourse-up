#!/usr/bin/env ruby

`cp binary-linux/concourse-up-linux-amd64 ./cup`
`chmod +x ./cup`

system_tests_name_prefix = "concourse-up-ts"

non_empty = []

buckets = `aws s3 ls | grep -E '#{system_tests_name_prefix}' | awk '{print $3}'`

if buckets.empty?
  puts 'No residual systest buckets found'
  exit(0)
end

Orphan = Struct.new(:bucket, :deployment, :project, :region)

buckets.each_line do |bucket|
  bucket = bucket.strip
  deployment = bucket.split('-')[0..4].join('-')
  project = bucket.split('-')[2..4].join('-')
  region = bucket.split('-')[5..7].join('-')
  orphan = Orphan.new(bucket, deployment, project, region)

  puts "Processing #{orphan}"

  aws_region_cmd = "aws --region #{orphan.region}"

  `aws s3 ls s3://#{orphan.bucket} > /dev/null`
  if $?.success?
    contents = `aws s3 ls s3://#{orphan.bucket}`
    if contents.include?('terraform.tfstate') && contents.include?('director-state.json')
      non_empty.push(orphan)
    else
      puts "#{orphan.bucket} is missing key files"
      puts 'Attempting to delete VPC'
      puts '==================================================================='
      puts "MANUAL CLEANUP MAY BE REQUIRED FOR #{orphan.bucket}"
      puts '==================================================================='

      vpc_id = `#{aws_region_cmd} ec2 describe-vpcs --filters 'Name=tag:Name,Values=#{orphan.project}' | jq -r '.Vpcs[0].VpcId'`
      unless vpc_id.empty? || vpc_id.include?('null')
        rds_instance = `#{aws_region_cmd} rds describe-db-instances | jq -r '.DBInstances[] | select(.DBSubnetGroup.VpcId == "#{vpc_id}") | .DBInstanceIdentifier'`
        if rds_instance
          `#{aws_region_cmd} rds delete-db-instance --db-instance-identifier "#{rds_instance}" --skip-final-snapshot true`
          `#{aws_region_cmd} rds wait db-instance-deleted --db-instance-identifier "#{rds_instance}"`
        end
        `#{aws_region_cmd} ec2 delete-vpc --vpc-id #{vpc_id}`
      else
        puts "vpc for #{orphan.deployment} not found"
      end
      `aws s3 rb s3://#{orphan.bucket} --force`
      puts "Deleted bucket #{orphan.bucket}"
    end
  else
    printf "#{orphan.bucket} doesn't really exist\n"
  end
end

exit(0) if non_empty.empty?

non_empty.each do |orphan|
  command = "./cup destroy --region #{orphan.region} #{orphan.project}"
  puts "Attempting to run #{command}"
  str = sprintf("yes yes | ./cup destroy --region %s %s", orphan.region, orphan.project)
  `echo '#{str}' >> to_delete`
end

# Do deletes in Bash so we can see STDOUT as it happens without needing any Gems
`set -x && \
 sort -u to_delete \
 | xargs -P 8 -I {} bash -c '{}'`

`rm -f to_delete`

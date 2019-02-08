#!/usr/bin/env ruby

`cp binary-linux/concourse-up-linux-amd64 ./cup`
`chmod +x ./cup`

class AWS
  private
    attr_reader :system_tests_name_prefix
  public

  def initialize(system_tests_name_prefix)
    @system_tests_name_prefix = system_tests_name_prefix
  end

  def bucket_names
    `aws s3 ls | grep -E '#{system_tests_name_prefix}' | awk '{print $3}'`.split("\n")
  end

  def new_orphan(bucket_name, deployment, project, region)
    AWSOrphan.new(AWSBucket.new(bucket_name), deployment, project, region)
  end
end

class AWSOrphan
  attr_reader :bucket, :deployment, :project, :region
  private
    attr_reader :aws_region_cmd
  public

  def initialize(bucket, deployment, project, region)
    @bucket = bucket
    @deployment = deployment
    @project = project
    @region = region
    aws_region_cmd = "aws --region #{region}"
  end

  def vpc_id
    `#{aws_region_cmd} ec2 describe-vpcs --filters 'Name=tag:Name,Values=#{orphan.project}' | jq -r '.Vpcs[0].VpcId'`
  end

  def vpc_id_valid?(id)
    !(id.empty? || id.include?('null'))
  end

  def delete_vpc(vpc_id)
    `#{aws_region_cmd} ec2 delete-vpc --vpc-id #{vpc_id}`
  end

  def db_instance_for(vpc_id)
    `#{aws_region_cmd} rds describe-db-instances | jq -r '.DBInstances[] | select(.DBSubnetGroup.VpcId == "#{vpc_id}") | .DBInstanceIdentifier'`
  end

  def db_instance_exists?(db_instance_id)
    db_instance_id
  end

  def delete_db_instance(vpc_id)
    db_instance_id = db_instance_for(vpc_id)
    if db_instance_exists?(db_instance_id)
      `#{aws_region_cmd} rds delete-db-instance --db-instance-identifier "#{rds_instance}" --skip-final-snapshot true`
      `#{aws_region_cmd} rds wait db-instance-deleted --db-instance-identifier "#{rds_instance}"`
    end
  end
end

class Cleaner
  private
    attr_reader :iaas
  public

  def initialize(iaas)
    @iaas = iaas
  end

  def orphans(buckets)
    buckets.map do |bucket_name|
      bucket_name = bucket_name.strip
      deployment = bucket_name.split('-')[0..4].join('-')
      project = bucket_name.split('-')[2..4].join('-')
      region = bucket_name.split('-')[5..7].join('-')

      iaas.new_orphan(bucket_name, deployment, project, region)
    end
  end

  def clean
    non_empty = []
    buckets = iaas.bucket_names

    if buckets.empty?
      puts 'No residual systest buckets found'
      exit(0)
    end

    orphans(buckets).each do |orphan|
      puts "Processing [#{orphan.bucket.name}]"
      bucket = orphan.bucket

      if bucket.exists?
        if bucket.key_files?
          non_empty.push(orphan)
        else
          puts "#{bucket.name} is missing key files"
          puts 'Attempting to delete VPC'
          puts '==================================================================='
          puts "MANUAL CLEANUP MAY BE REQUIRED FOR #{bucket.name}"
          puts '==================================================================='

          vpc_id = orphan.vpc_id
          if orphan.vpc_id_valid?
            orphan.delete_db_instance(vpc_id)
            orphan.delete_vpc(vpc_id)
          else
            puts "VPC for [#{orphan.deployment}] not found"
          end
          bucket.delete
          puts "Deleted bucket [#{bucket.name}]"
        end
      else
        printf "Bucket [#{bucket.name}] doesn't really exist\n"
      end
    end

    if non_empty.empty?
      return
    end

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
  end
end

class AWSBucket
  attr_reader :name

  def initialize(name)
    @name = name
  end

  def exists?
    `aws s3 ls s3://#{name} > /dev/null`
    $?.success?
  end

  def contents
    `aws s3 ls s3://#{name}`
  end

  def key_files?
    contents.include?('terraform.tfstate') && contents.include?('director-state.json')
  end

  def delete
    `aws s3 rb s3://#{name} --force`
  end
end

cleaner = Cleaner.new(AWS.new('concourse-up-ts'))
cleaner.clean

require 'json'

class AWS
  private
  attr_reader :system_tests_name_prefix

  public
  attr_reader :name

  def initialize(system_tests_name_prefix)
    @system_tests_name_prefix = system_tests_name_prefix
    @name = 'aws'
  end

  def bucket_names
    response_json = `aws s3api list-buckets`
    response = JSON.parse(response_json)
    response.fetch('Buckets')
            .map { |bucket| bucket.fetch('Name') }
            .select { |bucket_name| bucket_name.start_with?(system_tests_name_prefix) }
            .select { |bucket_name| bucket_name.end_with?('config') }
  end

  def new_orphan(bucket_name)
    deployment = bucket_name.split('-')[0..4].join('-')
    project = bucket_name.split('-')[2..4].join('-')
    region = bucket_name.split('-')[5..7].join('-')

    Orphan.new(AWSBucket.new(bucket_name), deployment, project, region)
  end

  def cleanup(orphan)
    bucket = orphan.bucket
    puts "\n#{bucket.name} is missing key files, attempting item-by-item delete"

    vpc_id = vpc_id(orphan.project, orphan.region)
    if vpc_id_valid?(vpc_id)
      threads = []

      threads << Thread.new { delete_bosh(orphan.project, orphan.region) }
      threads << Thread.new { delete_db_instances(vpc_id, orphan.region) }
      threads << Thread.new { delete_instances(orphan.project, orphan.region) }
      threads << Thread.new { delete_nat_gateways(orphan.project, orphan.region) }

      threads.each { |thread| thread.join }

      delete_elastic_ips(orphan.project, orphan.region)
      delete_subnets(orphan.project, orphan.region)
      delete_security_groups(orphan.project, orphan.region)
      delete_route_tables(orphan.project, orphan.region)
      delete_internet_gateways(orphan.project, orphan.region, vpc_id)
      delete_key_pairs(orphan.project, orphan.region)
      delete_network_acls(orphan.project, orphan.region, vpc_id)
      delete_vpc(vpc_id, orphan.region)
      if !$?.success?
        puts "Failed deleting VPC #{vpc_id} for bucket #{bucket.name}, skipping bucket delete"
        return
      end
    else
      puts "VPC for [#{orphan.deployment}] not found"
    end
    bucket.delete
    puts "Deleted bucket [#{bucket.name}]"
    blobstore_bucket = AWSBucket.new(bucket.name.sub!('config', 'blobstore'))
    blobstore_bucket.delete
    puts "Deleted bucket [#{blobstore_bucket.name}]"
  end

  private

  def vpc_id(project, region)
    results_json = run("aws --region #{region} ec2 describe-vpcs --filters 'Name=tag:concourse-up-project,Values=#{project}'")
    results = JSON.parse(results_json)
    return '' unless results.fetch('Vpcs', []).any?
    results.fetch('Vpcs')
           .first
           .fetch('VpcId')
  end

  def vpc_id_valid?(id)
    !(id.empty? || id.include?('null'))
  end

  def delete_vpc(vpc_id, region)
    run("aws --region #{region} ec2 delete-vpc --vpc-id #{vpc_id}")
  end

  def delete_db_instances(vpc_id, region)
    results_json = run("aws --region #{region} rds describe-db-instances")
    results = JSON.parse(results_json)
    ids = results.fetch('DBInstances')
                 .select { |db| db.fetch('DBSubnetGroup').fetch('VpcId') == vpc_id }
                 .map { |gateway| gateway.fetch('DBInstanceIdentifier') }
    ids.each do |id|
      run("aws --region #{region} rds delete-db-instance --db-instance-identifier '#{id}' --skip-final-snapshot")
      run("aws --region #{region} rds wait db-instance-deleted --db-instance-identifier '#{id}'")
    end
  end

  def delete_bosh(project, region)
    results_json = run("aws --region #{region} ec2 describe-instances --filter 'Name=tag:concourse-up-project,Values=#{project}' 'Name=tag:deployment,Values=bosh'")
    results = JSON.parse(results_json)
    ids = results.fetch('Reservations')
                .flat_map { |reservation| reservation.fetch('Instances') }
                .flat_map { |instance| instance.fetch('InstanceId') }
    ids.each { |id| run("aws --region #{region} ec2 terminate-instances --instance-id #{id}") }
  end

  def delete_instances(project, region)
    results_json = run("aws --region #{region} ec2 describe-instances --filter 'Name=tag:concourse-up-project,Values=#{project}'")
    results = JSON.parse(results_json)
    ids = results.fetch('Reservations')
                .flat_map { |reservation| reservation.fetch('Instances') }
                .flat_map { |instance| instance.fetch('InstanceId') }
    ids.each { |id| run("aws --region #{region} ec2 terminate-instances --instance-id #{id}") }
  end

  def delete_nat_gateways(project, region)
    delete_things('nat-gateway', 'NatGateways', 'NatGatewayId', 'nat-gateway-id', project, region)
    puts "NAT Gateway thread sleeping"
    sleep(60)
  end

  def delete_elastic_ips(project, region)
    results_json = run("aws --region #{region} ec2 describe-addresses --filter 'Name=tag:concourse-up-project,Values=#{project}'")
    results = JSON.parse(results_json)
    ids = results.fetch('Addresses').map { |address| address.fetch('AllocationId') }
    ids.each { |id| run("aws --region #{region} ec2 release-address --allocation-id #{id}") }
  end

  def delete_key_pairs(project, region)
    results_json = run("aws --region #{region} ec2 describe-key-pairs")
    results = JSON.parse(results_json)
    names = results.fetch('KeyPairs')
                   .map { |key_pairs| key_pairs.fetch('KeyName') }
                   .select { |key_name| key_name.start_with?(project) }
    names.each { |name| run("aws --region #{region} ec2 delete-key-pair --key-name #{name}") }
  end

  def delete_network_acls(project, region, vpc_id)
    results_json = run("aws --region #{region} ec2 describe-network-acls")
    results = JSON.parse(results_json)
    ids = results.fetch('NetworkAcls')
                   .select { |acl| acl.fetch('VpcId') == vpc_id }
                   .map { |acl| acl.fetch('NetworkAclId') }
    ids.each { |id| run("aws --region #{region} ec2 delete-network-acl --network-acl-id #{id}") }
  end

  def delete_volumes(project, region)
    delete_things('volume', 'Volumes', 'VolumeId', 'volume-id', project, region)
  end

  def delete_subnets(project, region)
    delete_things('subnet', 'Subnets', 'SubnetId', 'subnet-id', project, region)
  end

  def delete_security_groups(project, region)
    delete_things('security-group', 'SecurityGroups', 'GroupId', 'group-id', project, region)
  end

  def delete_route_tables(project, region)
    delete_things('route-table', 'RouteTables', 'RouteTableId', 'route-table-id', project, region)
  end

  def delete_internet_gateways(project, region, vpc_id)
    results_json = run("aws --region #{region} ec2 describe-internet-gateways --filter 'Name=tag:concourse-up-project,Values=#{project}'")
    results = JSON.parse(results_json)
    ids = results.fetch('InternetGateways').map { |address| address.fetch('InternetGatewayId') }
    ids.each { |id| run("aws --region #{region} ec2 detach-internet-gateway --vpc-id #{vpc_id} --internet-gateway-id #{id}") }
    ids.each { |id| run("aws --region #{region} ec2 delete-internet-gateway --internet-gateway-id #{id}") }
  end

  def delete_things(type, key, id_key, id_flag, project, region)
    results_json = run("aws --region #{region} ec2 describe-#{type}s --filter 'Name=tag:concourse-up-project,Values=#{project}'")
    results = JSON.parse(results_json)
    ids = results.fetch(key).map { |address| address.fetch(id_key) }
    ids.each { |id| run("aws --region #{region} ec2 delete-#{type} --#{id_flag} #{id}") }
  end

  def run(command)
    puts "Thread [#{Thread.current.object_id}] running: #{command}"
    `#{command}`
  end
end

# So requirers don't have to know which IAAS
IAAS = AWS

class AWSBucket
  attr_reader :name

  def initialize(name)
    @name = name
  end

  def exists?
    `aws s3 ls s3://#{name} > /dev/null`
    $?.success?
  end

  def not_exists?
    `aws s3 ls s3://#{name} > /dev/null`
    !($?.success?)
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
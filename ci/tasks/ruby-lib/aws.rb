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
    `aws s3 ls | grep -E '#{system_tests_name_prefix}' | awk '{print $3}'`.split("\n")
  end

  def new_orphan(bucket_name)
    bucket_name = bucket_name.strip
    deployment = bucket_name.split('-')[0..4].join('-')
    project = bucket_name.split('-')[2..4].join('-')
    region = bucket_name.split('-')[5..7].join('-')

    Orphan.new(AWSBucket.new(bucket_name), deployment, project, region)
  end

  def cleanup(orphan)
    bucket = orphan.bucket
    puts "#{bucket.name} is missing key files"
    puts 'Attempting to delete VPC'
    puts '==================================================================='
    puts "MANUAL CLEANUP MAY BE REQUIRED FOR #{bucket.name}"
    puts '==================================================================='

    vpc_id = vpc_id(orphan.project, orphan.region)
    if vpc_id_valid?(vpc_id)
      delete_nat_gateways(orphan.project)
      delete_db_instance(vpc_id, orphan.region)
      delete_vpc(vpc_id, orphan.region)
    else
      puts "VPC for [#{orphan.deployment}] not found"
    end
    bucket.delete
    puts "Deleted bucket [#{bucket.name}]"
  end

  private

  def vpc_id(project, region)
    `aws --region #{region} ec2 describe-vpcs --filters 'Name=tag:Name,Values=#{project}' | jq -r '.Vpcs[0].VpcId'`
  end

  def vpc_id_valid?(id)
    !(id.empty? || id.include?('null'))
  end

  def delete_vpc(vpc_id, region)
    `aws --region #{region} ec2 delete-vpc --vpc-id #{vpc_id}`
  end

  def db_instance_for(vpc_id, region)
    `aws --region #{region} rds describe-db-instances | jq -r '.DBInstances[] | select(.DBSubnetGroup.VpcId == "#{vpc_id}") | .DBInstanceIdentifier'`
  end

  def db_instance_exists?(db_instance_id)
    db_instance_id
  end

  def delete_db_instance(vpc_id, region)
    db_instance_id = db_instance_for(vpc_id, region)
    if db_instance_exists?(db_instance_id)
      `aws --region #{region} rds delete-db-instance --db-instance-identifier "#{rds_instance}" --skip-final-snapshot true`
      `aws --region #{region} rds wait db-instance-deleted --db-instance-identifier "#{rds_instance}"`
    end
  end

  def delete_nat_gateways(project)
    results_json = `aws ec2 describe-nat-gateways --filter 'Name=tag:concourse-up-project,Values=#{project}'`
    results = JSON.parse(results_json)
    ids = results.fetch('NatGateways').map { |gateway| gateway.fetch('NatGatewayId') }
    ids.each { |id| `aws ec2 delete-nat-gateway --nat-gateway-id #{id}` }
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
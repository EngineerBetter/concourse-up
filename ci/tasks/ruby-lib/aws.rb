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
    `#{aws_region_cmd} ec2 describe-vpcs --filters 'Name=tag:Name,Values=#{project}' | jq -r '.Vpcs[0].VpcId'`
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
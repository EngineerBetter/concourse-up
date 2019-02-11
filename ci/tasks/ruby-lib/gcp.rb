require 'json'

pwd = `echo $PWD`.strip
creds_path = pwd+'googlecreds.json'
File.write(creds_path, ENV.fetch('GOOGLE_APPLICATION_CREDENTIALS_CONTENTS'))
ENV['GOOGLE_APPLICATION_CREDENTIALS'] = creds_path
`gcloud auth activate-service-account --key-file="$GOOGLE_APPLICATION_CREDENTIALS"`
ENV['CLOUDSDK_CORE_PROJECT'] = 'concourse-up'

class GCP
  private
  attr_reader :system_tests_name_prefix

  public
  attr_reader :name

  def initialize(system_tests_name_prefix)
    @system_tests_name_prefix = system_tests_name_prefix
    @name = 'gcp'
  end

  def bucket_names
    output = `gsutil ls gs://`
    output
      .split("\n")
      .map { |url| url.strip.delete_prefix('gs://').chomp('/') }
      .select { |url| url.start_with?(system_tests_name_prefix) }
  end

  def new_orphan(bucket_name)
    bucket_name = bucket_name.strip
    deployment = bucket_name.split('-')[0..4].join('-')
    project = bucket_name.split('-')[2..4].join('-')
    region = bucket_name.split('-')[5..6].join('-')

    Orphan.new(GCPBucket.new(bucket_name), deployment, project, region)
  end

  def cleanup(orphan)
    bucket = orphan.bucket
    if bucket.contents.empty?
      bucket.delete_empty
    else
      found_components = 0
      found_components += delete_regional_components('compute instances', "labels.deployment:bosh AND labels.concourse-up-project:#{orphan.project}", 'name', orphan.region)
      found_components += delete_regional_components('compute instances', "labels.concourse-up-project:#{orphan.project}", 'name', orphan.region)
      found_components += delete_regional_components('compute instances', "name:'#{orphan.deployment}-nat-instance'", 'name', orphan.region)
      found_components += delete_regional_components('compute networks subnets', "name~'.*#{orphan.project}.*'", 'name', orphan.region)
      found_components += delete_components('compute routes', "network~'.*#{orphan.deployment}.*'", 'name')
      found_components += delete_components('compute firewall-rules', "name~'.*#{orphan.project}.*'", 'name')
      found_components += delete_components('compute networks', "name~'.*#{orphan.project}.*'", 'name')
      found_components += delete_components('sql instances', "name~'.*#{orphan.project}.*'", 'name')
      found_components += delete_components('iam service-accounts', "email~'.*#{orphan.project}.*'", 'uniqueId')
      found_components += delete_components('compute addresses', "name~'.*#{orphan.project}.*'", 'name')

      # Only delete the bucket if there are no orphaned components. This will never be true on the first run.
      # We need to do this, rather than detecting failures to delete, as we don't know _why_ the deployment
      # failed and so half of the components could have been missing to start with.
      if found_components == 0
        bucket.delete
      end
    end
  end

  private

  def delete_components(component_type, filter, identifier_key, region_flag = "")
    componentsJson = `gcloud --format json #{component_type} list --filter="#{filter}"`
    components = JSON.parse(componentsJson)
    componentIdentifiers = components.map { |component| component.fetch("#{identifier_key}") }
    componentIdentifiers.each { |componentIdentifier| `gcloud --quiet #{region_flag} #{component_type} delete #{componentIdentifier}` }
    components.size
  end

  def delete_regional_component(component_type, filter, identifier_key, region)
    delete_components(component_type, filter, identifier_key, "--region=#{region}")
  end
end

# So requirers don't have to know which IAAS
IAAS = GCP

class GCPBucket
  attr_reader :name

  def initialize(name)
    @name = name
  end

  def exists?
    `gsutil ls gs://#{name} > /dev/null`
    $?.success?
  end

  def not_exists?
    !exists?
  end

  def contents
    `gsutil ls gs://#{name}`
  end

  def key_files?
    contents.include?('default.tfstate') && contents.include?('director-state.json')
  end

  def delete
    `gsutil rm -r gs://#{name}`
  end

  def delete_empty
    `gsutil rb gs://#{name}`
  end
end

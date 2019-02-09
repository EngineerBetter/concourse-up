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
      puts 'full cleanup for deployments without key Terraform and state files not supported on GCP'
    end
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
    puts "Would run gsutil rm -r gs://#{name}"
    # `gsutil rm -r gs://#{name}`
  end

  def delete_empty
    `gsutil rb gs://#{name}`
  end
end
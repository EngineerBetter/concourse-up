#!/usr/bin/env ruby

require_relative('ruby-lib/aws.rb')
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
    destroyable_orphans = []
    buckets = iaas.bucket_names

    if buckets.empty?
      puts 'No residual systest buckets found'
      exit(0)
    end

    orphans(buckets).each do |orphan|
      puts "Processing [#{orphan.bucket.name}]"
      bucket = orphan.bucket

      if bucket.not_exists?
        printf "Bucket [#{bucket.name}] doesn't really exist\n"
        next
      end

      if bucket.key_files?
        destroyable_orphans.push(orphan)
      else
        cleanup_via_iaas(orphan)
      end
    end

    cleanup_with_concourse_up(destroyable_orphans) unless destroyable_orphans.empty?
  end

  def cleanup_via_iaas(orphan)
    bucket = orphan.bucket
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

  def cleanup_with_concourse_up(orphans_with_state_files)
    orphans_with_state_files.each do |orphan|
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

`cp binary-linux/concourse-up-linux-amd64 ./cup`
`chmod +x ./cup`

cleaner = Cleaner.new(AWS.new('concourse-up-ts'))
cleaner.clean

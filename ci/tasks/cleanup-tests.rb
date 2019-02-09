#!/usr/bin/env ruby

require_relative("ruby-lib/#{ENV.fetch('IAAS').downcase}.rb")

class Orphan
  attr_reader :bucket, :deployment, :project, :region

  def initialize(bucket, deployment, project, region)
    @bucket = bucket
    @deployment = deployment
    @project = project
    @region = region
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
      iaas.new_orphan(bucket_name)
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
        iaas.cleanup(orphan)
      end
    end

    cleanup_with_concourse_up(destroyable_orphans) unless destroyable_orphans.empty?
  end

  def cleanup_with_concourse_up(orphans_with_state_files)
    orphans_with_state_files.each do |orphan|
      command = "./cup destroy --region #{orphan.region} #{orphan.project}"
      puts "Attempting to run #{command}"
      str = sprintf("yes yes | ./cup destroy --iaas #{iaas.name} --region %s %s", orphan.region, orphan.project)
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

cleaner = Cleaner.new(IAAS.new('concourse-up-ts'))
cleaner.clean

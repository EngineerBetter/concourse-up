#!/usr/bin/env ruby

require 'json'
require 'erb'

d = {}
o = {}
if STDIN.tty?
  puts 'cat file | get-versions'
else
  data = JSON.parse(STDIN.read)
  data.each do |e|
    if e['path'].include?('stemcells')
      d['aws-xenial-stemcell'] = e['value']
    else
      d["#{e['value']['name']}-release"] = e['value']['version']
    end
  end
  d.each do |k, v|
    filename = "#{k}/version"
    if File.exist? filename
      version = File.open(filename).read.strip
      if version != v
        o[k] = version
      end
    else
      puts "filename #{filename} does not exist"
    end
  end
end

release = File.read("release/version").strip

output = ERB.new(<<-BLOCK).result(binding)

Release <%= release %> of Concourse-Up is now available. 
<% if !o.empty? %>
The changes since last release are:
<% o.each do |k,v| %>
- <%= k.sub("-release", "").capitalize %> <%= d[k] %> --> <%= k.sub("-release", "").capitalize %> <%= v %>
<% end %>
<% end %>
BLOCK

puts output

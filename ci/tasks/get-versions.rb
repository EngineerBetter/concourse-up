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
      if version.split('.').first != v.split('.').first
        o[k] = version
      end
    else
      puts "filename #{filename} does not exist"
    end
  end
end

if !o.empty?
  output = ERB.new(<<-BLOCK).result(binding)
<% o.each do |k,v| %>
  Name: <%= k %>
  From: <%= d[k] %>
  To: <%= v %>
<% end %>
BLOCK
  puts output
end
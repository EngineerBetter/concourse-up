#!/usr/bin/env ruby

require 'json'
require 'erb'
require 'net/http'
require 'uri'

def fetch(uri_str, limit = 10)
    # You should choose a better exception.
    raise ArgumentError, 'too many HTTP redirects' if limit == 0

    url = URI.parse(uri_str)
    req = Net::HTTP::Get.new(url.request_uri)
    req['Accept'] = "application/vnd.github.v3+json"

    http = Net::HTTP.new(url.host, url.port)
    http.use_ssl = (url.scheme == "https")
    res = http.request(req)

    case res
    when Net::HTTPSuccess then
      res.body
    when Net::HTTPRedirection then
      location = res['location']
      fetch(location, limit - 1)
    else
      res.value
    end
end

  
fetched=fetch("https://api.github.com/repos/EngineerBetter/concourse-up/releases\?prerelease\=false")

a=JSON.parse(fetched)
h={}

current_release = a[0]["url"]
previous_release = a[7]["url"]

current_ops_link = fetch(current_release).match(/https:\/\/github\.com\/EngineerBetter\/concourse-up-ops\/tree\/\d+\.\d+\.\d+/i).to_s.gsub('tree','raw')
previous_ops_link = fetch(previous_release).match(/https:\/\/github\.com\/EngineerBetter\/concourse-up-ops\/tree\/\d+\.\d+\.\d+/i).to_s.gsub('tree','raw')

current_ops_file = fetch("#{current_ops_link}/ops/versions.json")
previous_ops_file = fetch("#{previous_ops_link}/ops/versions.json")

h[0] = JSON.parse(current_ops_file)
h[1] = JSON.parse(previous_ops_file)

found={}
h[0].each do |f|
    version0=""
    version1=""
    name="" 
    if f["path"] == "/stemcells/alias=xenial/version"
        version0=f["value"]
        e=h[1].detect {|v| v["path"] == "/stemcells/alias=xenial/version"}
        version1=e["value"]
        name="stemcell"
    else
        version0=f["value"]["version"]
        name=f["value"]["name"]
        e=h[1].detect {|v| v["value"]["name"] == name}
        version1=e["value"]["version"]
    end

    if version0 != version1
        found[name]={version0: version0, version1: version1}
    end
end

if found.length > 0
    puts "The changes since last release are:"
    found.each do |k, v|
        puts "- #{k} #{v[:version1]} --> #{k} #{v[:version0]}"
    end
end



require 'yaml'
require 'aws-sdk'

RELEASE_PATH = ENV.fetch("RELEASE_PATH")

def get_blob_id_by_prefix(prefix)
  reader = File.open(File.join(RELEASE_PATH, 'config/blobs.yml'))
  objs = YAML.load(reader)
  name, attrs = objs.detect {|key, obj| key.include? prefix}
  attrs['object_id']
end

def cred
  reader = File.open(File.join(RELEASE_PATH, 'config/final.yml'))
  objs = YAML.load(reader)
  objs['blobstore']['options']
end

def s3
  Aws::S3::Client.new(
    access_key_id: cred['access_key_id'],
    secret_access_key: cred['secret_access_key'],
    region: "us-east-1"
  )
end

def download_blob(obj_id, filename)
  File.open(filename, 'wb') do |file|
    s3.get_object(bucket: cred['bucket_name'], key: obj_id) do |chunk|
      file.write(chunk)
    end
  end
end

clamav_obj_id = get_blob_id_by_prefix('clamav/clamav-')
pcre_obj_id = get_blob_id_by_prefix('pcre2/pcre2-')

download_blob clamav_obj_id, 'clamav-blob.tar.gz'
download_blob pcre_obj_id, 'pcre2-blob.tar.gz'

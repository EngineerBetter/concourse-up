mkdir -p $GOPATH/src/bitbucket.org/engineerbetter/concourse-up
mv concourse-up/* $GOPATH/src/bitbucket.org/engineerbetter/concourse-up
cd $GOPATH/src/bitbucket.org/engineerbetter/concourse-up

if [ -z ${AWS_ACCESS_KEY_ID+x} ]
then
  mkdir ~/.aws
  cat > ~/.aws/config <<- EOF
[default]
aws_access_key_id="${AWS_ACCESS_KEY_ID}"
aws_secret_access_key="${AWS_SECRET_ACCESS_KEY}"
region=eu-west-1
output=json
EOF
fi
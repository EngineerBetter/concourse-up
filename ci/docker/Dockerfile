FROM golang:1.8
RUN apt-get update && apt-get install -y --no-install-recommends \
  build-essential \
  libreadline6 \
  libreadline6-dev \
  libsqlite3-dev \
  libssl-dev \
  libxml2-dev \
  libxslt-dev \
  libyaml-dev \
  openssl \
  python-dev \
  python-pip \
  ruby \
  ruby-dev \
  sqlite3 \
  unzip \
  zlib1g-dev \
  zlibc \
  vim
RUN pip install --upgrade pip
RUN pip install --upgrade virtualenv
RUN pip install awscli

RUN gem install bosh_cli --no-doc --no-ri

RUN go get github.com/onsi/ginkgo/ginkgo github.com/onsi/gomega \
           gopkg.in/alecthomas/gometalinter.v1

RUN gometalinter.v1 --install

RUN wget https://releases.hashicorp.com/terraform/0.9.3/terraform_0.9.3_linux_amd64.zip \
  && unzip terraform_0.9.3_linux_amd64.zip \
  && chmod +x terraform \
  && mv terraform /usr/bin/terraform

RUN wget https://s3.amazonaws.com/bosh-init-artifacts/bosh-init-0.0.102-linux-amd64 \
  && mv bosh-init-0.0.102-linux-amd64 /usr/bin/bosh-init \
  && chmod +x /usr/bin/bosh-init

RUN wget https://github.com/stedolan/jq/releases/download/jq-1.5/jq-linux64 \
    && mv jq-linux64 /usr/bin/jq \
    && chmod +x /usr/bin/jq


COPY verify_image.sh /tmp/verify_image.sh
RUN /tmp/verify_image.sh && rm /tmp/verify_image.sh
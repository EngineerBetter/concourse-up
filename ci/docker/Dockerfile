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
    vim \
    zlib1g-dev \
    zlibc

RUN pip install --upgrade pip \
    && pip install --upgrade virtualenv \
    && pip install awscli

RUN go get github.com/onsi/ginkgo/ginkgo \
           github.com/onsi/gomega \
           gopkg.in/alecthomas/gometalinter.v1 \
           github.com/square/certstrap

RUN gometalinter.v1 --install

RUN wget https://github.com/stedolan/jq/releases/download/jq-1.5/jq-linux64 \
    && mv jq-linux64 /usr/bin/jq \
    && chmod +x /usr/bin/jq

RUN wget https://github.com/concourse/concourse/releases/download/v3.4.1/fly_linux_amd64 \
    && mv fly_linux_amd64 /usr/bin/fly \
    && chmod +x /usr/bin/fly

RUN wget https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-2.0.28-linux-amd64 \
    && mv bosh-cli-2.0.28-linux-amd64 /usr/bin/bosh-cli \
    && chmod +x /usr/bin/bosh-cli

COPY verify_image.sh /tmp/verify_image.sh
RUN /tmp/verify_image.sh && rm /tmp/verify_image.sh

FROM golang:1
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    default-libmysqlclient-dev \
    libpq-dev \
    libsqlite3-dev \
    libssl-dev \
    libxml2-dev \
    libxslt1-dev \
    python-dev \
    python-pip \
    python-setuptools \
    ruby \
    ruby-dev \
    unzip \
    zlib1g-dev

RUN pip install --upgrade pip \
    && pip install --upgrade virtualenv \
    && pip install awscli

RUN go get github.com/onsi/ginkgo/ginkgo \
    github.com/onsi/gomega \
    gopkg.in/alecthomas/gometalinter \
    github.com/square/certstrap

RUN gometalinter --install

RUN wget https://github.com/stedolan/jq/releases/download/jq-1.5/jq-linux64 \
    && mv jq-linux64 /usr/bin/jq \
    && chmod +x /usr/bin/jq

RUN wget https://github.com/concourse/concourse/releases/download/v3.4.1/fly_linux_amd64 \
    && mv fly_linux_amd64 /usr/bin/fly \
    && chmod +x /usr/bin/fly

RUN wget https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-2.0.28-linux-amd64 \
    && mv bosh-cli-2.0.28-linux-amd64 /usr/bin/bosh-cli \
    && chmod +x /usr/bin/bosh-cli

RUN wget http://ftp.uk.debian.org/debian/pool/main/s/shellcheck/shellcheck_0.4.6-1_amd64.deb \
    && dpkg -i shellcheck_0.4.6-1_amd64.deb \
    && rm shellcheck_0.4.6-1_amd64.deb

COPY verify_image.sh /tmp/verify_image.sh
RUN /tmp/verify_image.sh && rm /tmp/verify_image.sh

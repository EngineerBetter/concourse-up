FROM golang:1

RUN apt-get update && apt-get install -y --no-install-recommends \
    ruby \
    unzip \
    shellcheck \
    jq \
    awscli

RUN go get github.com/onsi/ginkgo/ginkgo \
    github.com/onsi/gomega \
    github.com/alecthomas/gometalinter \
    github.com/square/certstrap

RUN gometalinter --install

RUN wget https://github.com/concourse/concourse/releases/download/v3.14.1/fly_linux_amd64 \
    && mv fly_linux_amd64 /usr/bin/fly \
    && chmod +x /usr/bin/fly

RUN wget https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-5.0.1-linux-amd64 \
    && mv bosh-cli-5.0.1-linux-amd64 /usr/bin/bosh-cli \
&& chmod +x /usr/bin/bosh-cli

COPY verify_image.sh /tmp/verify_image.sh
RUN /tmp/verify_image.sh && rm /tmp/verify_image.sh

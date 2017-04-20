FROM golang:1.8
RUN apt-get update && apt-get install -y --no-install-recommends \
  python-pip \
  python-dev \
  build-essential \
  unzip \
  vim
RUN pip install --upgrade pip
RUN pip install --upgrade virtualenv
RUN pip install awscli

RUN go get github.com/onsi/ginkgo/ginkgo github.com/onsi/gomega \
           gopkg.in/alecthomas/gometalinter.v1

RUN wget https://releases.hashicorp.com/terraform/0.9.3/terraform_0.9.3_linux_amd64.zip \
  && unzip terraform_0.9.3_linux_amd64.zip \
  && chmod +x terraform \
  && mv terraform /usr/bin/terraform

RUN wget https://github.com/stedolan/jq/releases/download/jq-1.5/jq-linux64 \
    && mv jq-linux64 /usr/bin/jq \
    && chmod +x /usr/bin/jq

COPY verify_image.sh /tmp/verify_image.sh
RUN /tmp/verify_image.sh && rm /tmp/verify_image.sh
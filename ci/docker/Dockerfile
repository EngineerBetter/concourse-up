FROM golang:1.8
RUN apt-get update && apt-get install -y --no-install-recommends \
  python-pip \
  python-dev \
  build-essential \
  unzip
RUN pip install --upgrade pip
RUN pip install --upgrade virtualenv
RUN pip install awscli

RUN wget https://releases.hashicorp.com/terraform/0.9.3/terraform_0.9.3_linux_amd64.zip \
  && unzip terraform_0.9.3_linux_amd64.zip \
  && chmod +x terraform \
  && mv terraform /usr/bin/terraform

COPY verify_image.sh /tmp/verify_image.sh
RUN /tmp/verify_image.sh && rm /tmp/verify_image.sh
FROM golang:1.8
RUN apt-get update && apt-get install -y --no-install-recommends \
  python-pip \
  python-dev \
  build-essential
RUN pip install --upgrade pip
RUN pip install --upgrade virtualenv
RUN pip install awscli

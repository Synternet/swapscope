#
# 1. Build Container
#
FROM golang:1.20 AS build
ARG VERSION CI_JOB_TOKEN CI_SERVER_HOST NATS_HOSTS NATS_CREDS NATS_PUB_JWT NATS_PUB_NKEY PUB_SUBJECT_PREFIX

COPY . /home/src
WORKDIR /home/src

# Build components.
RUN make build

#
# 2. Runtime Container
#
FROM alpine

ENV TZ=Europe/Vilnius \
    PATH="/app:${PATH}"

RUN apk add --update --no-cache \
    tzdata \
    ca-certificates \
    bash \
    && \
    cp --remove-destination /usr/share/zoneinfo/${TZ} /etc/localtime && \
    echo "${TZ}" > /etc/timezone

WORKDIR /home/app
COPY --from=build /home/src/dist/swapscope .

# Required ENV vars to be set at runtime:
# - NATS_URL - NATS url
# - NATS_PUB_JWT - NATS JWT
# - NATS_PUB_NKEY - NATS NKEY
# - ETH_NODE - ETH node url
# - DB_HOST - DB host
# - DB_PORT - DB port
# - DB_USER - DB user
# - DB_PASSWORD - DB password
# - DB_NAME - DB name

CMD ["sh", "-c", "./swapscope"]
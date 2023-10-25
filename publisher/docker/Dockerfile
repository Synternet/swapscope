#
# 1. Build Container
#
FROM golang:1.20 AS build
ARG VERSION CI_JOB_TOKEN CI_SERVER_HOST NATS_HOSTS NATS_CREDS NATS_JWT NATS_NKEY PUB_SUBJECT_PREFIX

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

CMD ["sh", "-c", "./swapscope --nats ${NATS_URL} --nats-sub-jwt ${NATS_SUB_JWT} --nats-sub-nkey ${NATS_SUB_NKEY}"]
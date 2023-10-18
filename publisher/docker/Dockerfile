#
# 1. Build Container
#
FROM golang:1.20 AS build
ARG VERSION CI_JOB_TOKEN CI_SERVER_HOST NATS_HOSTS NATS_CREDS NATS_JWT NATS_NKEY PUB_SUBJECT_PREFIX

RUN bash -c 'echo -e "machine gitlab.com\nlogin gitlab-ci-token\npassword ${CI_JOB_TOKEN}" > ~/.netrc'
RUN git config --global url.https://gitlab-ci-token:${CI_JOB_TOKEN}@${CI_SERVER_HOST}.insteadOf https://${CI_SERVER_HOST}

COPY . /home/src
WORKDIR /home/src

# Build components.
RUN CGO_ENABLED=0 go build -v -o . -installsuffix cgo -ldflags="-w -s" ./cmd/liquial

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
COPY --from=build /home/src/liquial .

CMD ["sh", "-c", "./liquial --nats ${NATS_HOSTS} --nats-jwt ${NATS_JWT} --nats-nkey ${NATS_NKEY} --pub-subject-prefix ${PUB_SUBJECT_PREFIX}"]
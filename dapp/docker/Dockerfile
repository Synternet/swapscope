FROM node:16-alpine as build-stage

RUN apk add --no-cache git

ARG NEXT_PUBLIC_ENV
ARG NEXT_PUBLIC_NATS_URL
ARG NEXT_PUBLIC_ACCESS_TOKEN

RUN export NEXT_PUBLIC_ENV=$NEXT_PUBLIC_ENV
RUN export NEXT_PUBLIC_NATS_URL=$NEXT_PUBLIC_NATS_URL
RUN export NEXT_PUBLIC_ACCESS_TOKEN=$NEXT_PUBLIC_ACCESS_TOKEN

WORKDIR /app
COPY ./src ./src
COPY ./public ./public
COPY .eslintrc.json .gitignore .ts-prunerc.json .env \
    next.config.js package-lock.json package.json tsconfig.json ./
RUN npm i --omit=optional --legacy-peer-deps
RUN npm run check:ci
RUN npm run build
RUN npm run export

FROM nginx:1.23.3-alpine
COPY --from=build-stage /app/out /usr/share/nginx/html
COPY ./nginx.conf /etc/nginx/conf.d/default.conf
CMD [ "nginx","-g","daemon off;" ]
EXPOSE 80


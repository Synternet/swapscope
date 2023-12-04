# SwapScope Dapp
![NodeJS](https://img.shields.io/badge/node.js-6DA55F?style=for-the-badge&logo=node.js&logoColor=white)
![NPM](https://img.shields.io/badge/NPM-%23CB3837.svg?style=for-the-badge&logo=npm&logoColor=white)
![Docker](https://img.shields.io/badge/docker-%230db7ed.svg?style=for-the-badge&logo=docker&logoColor=white)

## Getting Started

```
nvm use                     - use correct npm version
npm i --legacy-peer-deps    - install packages ignoring resolution warnings 
npm run dev                 - run development server (http://localhost:3000)
npm run serve               - run production build locally (http://localhost:3000)
npm run dev:mock            - run development server with mocked api (http://localhost:3000)
npm run serve:mock          - run production build locally with mocked api (http://localhost:3000)
npm run cypress:mac         - run cypress tests on mac
```

## Docker

```
docker build -f ./docker/Dockerfile . -t swapscope      - create image
docker run -p 3000:80 -td swapscope                     - run container (http://localhost:3000)
```
   
## Env variables

```
NEXT_PUBLIC_NATS_URL        - url to nats (default is set to devnet)
NEXT_PUBLIC_ACCESS_TOKEN    - access token for your project (retrieve it from developer portal)
NEXT_PUBLIC_ENV             - dapp environment (local/prod)
NEXT_PUBLIC_SUBJECT_NAME    - subject to subscribe (stream in developer portal)
```
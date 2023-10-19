## Getting Started


```bash
nvm use                     - use correct npm version
npm i --legacy-peer-deps    - install packages ignoring resolution warnings 
npm run dev                 - run development server (http://localhost:3000)
npm run serve               - run production build locally (http://localhost:3000)
npm run dev:mock            - run development server with mocked api (http://localhost:3000)
npm run serve:mock          - run production build locally with mocked api (http://localhost:3000)
npm run cypress:mac         - run cypress tests on mac
```

## Docker

```bash
docker build -f ./docker/Dockerfile . -t swapscope      - create image
docker run -p 3000:80 -td swapscope                     - run container (http://localhost:3000)
```
   
## Env variables

```bash
NEXT_PUBLIC_NATS_URL        - url to nats (default is set to devnet)
NEXT_PUBLIC_ACCESS_TOKEN    - access token for your project (retrieve it from developer portal) 

```
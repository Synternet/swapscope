## Getting Started


```bash
nvm use                     - use correct npm version
npm i --legacy-peer-deps    - install packages ignoring resolution warnings 
npm run dev                 - run development server (http://localhost:3000)
npm run serve               - test production build locally (http://localhost:3000)
npm run cypress:mac         - run cypress tests on mac
```

## Docker

```bash
docker build -f ./docker/Dockerfile . -t swapscope        - create image
docker run -p 3000:80 -td swapscope - run container (http://localhost:3000)
```
FROM node:20.11.1-bookworm
WORKDIR /app/
COPY src /app/src
COPY prisma /app/prisma
COPY docker-entrypoint.sh /app/
COPY tsconfig.json /app/
COPY package.json /app/
COPY package-lock.json /app/
COPY srcpatch.sh /app/
RUN sh srcpatch.sh
RUN npm install --save-dev
RUN npm run build
RUN rm srcpatch.sh out/**/*.ts out/**/*.map
RUN rm -rf src
RUN npm prune --production
ENTRYPOINT sh docker-entrypoint.sh
FROM node:18-alpine

WORKDIR /app

RUN apk add --no-cache python3 make g++ linux-headers eudev-dev

COPY package.json ./
RUN npm install --production

COPY . .

RUN mkdir -p /app/logs /app/data/templates

EXPOSE 3001 3002

CMD ["npm", "start"]

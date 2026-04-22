FROM node:20-alpine
WORKDIR /app
RUN npm install -g expo-cli eas-cli
COPY package.json package-lock.json* ./
RUN npm ci
COPY . .
CMD ["expo", "export", "--platform", "web"]

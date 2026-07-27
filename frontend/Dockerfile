# Vite bakes VITE_* variables in at BUILD time, not at container start time,
# so they're passed in as build args (wired up in docker-compose.yml) rather
# than read from a runtime .env file inside the container.
ARG VITE_API_URL=http://localhost:8080
ARG VITE_BUSINESS_ID=default_business
ARG VITE_BUSINESS_NAME=Peak Mobile

FROM node:20-alpine AS builder
WORKDIR /app
ARG VITE_API_URL
ARG VITE_BUSINESS_ID
ARG VITE_BUSINESS_NAME
ENV VITE_API_URL=$VITE_API_URL
ENV VITE_BUSINESS_ID=$VITE_BUSINESS_ID
ENV VITE_BUSINESS_NAME=$VITE_BUSINESS_NAME

COPY package.json ./
RUN npm install
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]

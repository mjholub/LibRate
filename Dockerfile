FROM golang:1.20-alpine AS app

FROM node:alpine AS fronend

ADD . .

RUN cd fe && npm install && npm run build
RUN go mod tidy && go build

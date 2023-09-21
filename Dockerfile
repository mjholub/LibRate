FROM golang:1.21-alpine AS app

RUN addgroup -S librate && adduser -S librate -G librate

WORKDIR /app

COPY . .

RUN apk add --no-cache \
  nodejs-lts \
  npm 

RUN chown -R librate:librate /app
USER librate 
RUN cd fe && npm install && npm run build
RUN go mod tidy && go build

RUN ./LibRate -init && ./LibRate migrate -auto -exit

CMD ["./LibRate"]

EXPOSE 3000

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

# initialize the database, don't launch the database subprocess and rely solely on pg_isready, run the migrations and exit
RUN ./LibRate -init -no-db-subprocess -hc-extern && ./LibRate migrate -auto -exit -no-db-subprocess -hc-extern

CMD ["./LibRate", "-no-db-subprocess", "-hc-extern"]

EXPOSE 3000

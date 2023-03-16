FROM docker.io/library/golang:1.20-bullseye

WORKDIR /src

# copy go.mod, go.sum into /src,
# download dependencies
COPY go.* .
RUN go mod download

# copy all files into /src
COPY . ./

# build the application from source
RUN go build -o /media-rating-api

EXPOSE 8080

CMD [ "/media-rating-api" ]
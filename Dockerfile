ARG GO_VERSION=1.19

ARG TARGETARCH
 
###########
## Build ##
###########
FROM golang:${GO_VERSION} AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY src/ ./src/

# update swagger docs
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init

# build binary
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o /chartinstaller 

############
## Deploy ##
############
# FROM golang:${GO_VERSION}-alpine AS final
FROM scratch as final

RUN mkdir /helm && chown 1000 /helm && chgrp 1000 /helm
USER 1000:1000

COPY --from=build /app/docs/ /docs/
COPY --from=build /chartinstaller /chartinstaller

RUN mkdir /helm/data
RUN mkdir /helm/config
RUN mkdir /helm/temp

ENV GIN_MODE=release
ENV TARGET_NAMESPACE=default
ENV CHART_MUSEUM_URI=http://chartmuseum:8080
ENV HELM_DATA_HOME=/helm/data
ENV HELM_CONFIG_HOME=/helm/config
ENV HELM_CACHE_HOME=/helm/temp

ENV PORT=8080

EXPOSE 8080

CMD ["/chartinstaller"]

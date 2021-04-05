FROM golang AS build

ENV DISTRIBUTION_DIR /go/src/github.com/rolinux/rabbitmq-to-transmissionrpc

RUN apt-get update && apt-get install -y --no-install-recommends \
		git \
	&& rm -rf /var/lib/apt/lists/*

WORKDIR $DISTRIBUTION_DIR
COPY . $DISTRIBUTION_DIR

RUN go mod download

RUN CGO_ENABLED=0 go build -v -a -installsuffix cgo -o rabbitmq-to-transmissionrpc main.go

# run container with app on top on scratch empty container
FROM scratch

COPY --from=build /go/src/github.com/rolinux/rabbitmq-to-transmissionrpc/rabbitmq-to-transmissionrpc /bin/rabbitmq-to-transmissionrpc

CMD ["rabbitmq-to-transmissionrpc"]

FROM public.ecr.aws/docker/library/golang:1.18 as deps

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPRIVATE=github.com/saltpay/*

RUN git config --global url."git@github.com:".insteadOf "https://github.com/"

WORKDIR /app
COPY go.mod .
COPY go.sum .

# Copy the code into the container
COPY main.go .

FROM deps as build

# Build the application and copy somewhere convienient
RUN go build -o wait-for-kafka .

# create our new image with just the stuff we need
FROM public.ecr.aws/docker/library/alpine:3.14
WORKDIR /app
COPY --from=build /app/ ./

CMD ["/app/wait-for-kafka"]

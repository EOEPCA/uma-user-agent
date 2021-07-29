FROM golang:alpine as builder
WORKDIR /appbuild
COPY . .
RUN go build ./src/uma-user-agent

FROM alpine
WORKDIR /app
COPY --from=builder /appbuild/uma-user-agent .
ENTRYPOINT [ "/app/uma-user-agent" ]
CMD []

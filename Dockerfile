FROM golang:1.20.1-alpine3.16 AS builder

RUN apk update && \
  apk add --no-cache --update make bash git ca-certificates && \
  update-ca-certificates

WORKDIR /go/src/lineq-operator

COPY . .

RUN make build

FROM alpine:3.16.0

COPY --from=builder /go/src/lineq-operator/bin/lineq-operator /lineq-operator

CMD [ "/lineq-operator" ]
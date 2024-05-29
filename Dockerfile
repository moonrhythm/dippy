FROM golang:1.22.3

ENV CGO_ENABLED=0
WORKDIR /workspace
ADD . .
RUN go build -o dippy .

FROM gcr.io/distroless/static

COPY --from=0 /workspace/dippy .

ENTRYPOINT ["/dippy"]

FROM golang:1.20.3

ENV GOOS=linux
ENV CGO_ENABLED=0
WORKDIR /workspace
ADD . .
RUN go build -o dippy .

FROM gcr.io/distroless/static

COPY --from=0 /workspace/dippy .

ENTRYPOINT ["/dippy"]

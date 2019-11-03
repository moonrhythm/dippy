FROM golang:alpine as build

ENV GOOS=linux
ENV CGO_ENABLED=0
WORKDIR /workspace
ADD . .
RUN go build -o dippy .

FROM scratch
COPY --from=build /workspace/dippy .

ENTRYPOINT ["/dippy"]

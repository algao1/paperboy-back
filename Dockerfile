FROM golang:1.16-alpine3.13 as build
WORKDIR /go/src/github/paperboy/paperboy-back
COPY . .
RUN go build -o /paperboy-back ./cmd

FROM alpine:3.13.2
# COPY .env .
COPY --from=build /paperboy-back .

# ENV CACHE_URL host.docker.internal
# ENV CACHE_PORT 6379
# ENV PORT 8080

# EXPOSE 8080

CMD ["./paperboy-back"]
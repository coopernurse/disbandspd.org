# build stage
FROM golang:alpine AS build-env
ADD . /src
RUN cd /src && go build -o disbandspdweb cmd/disbandspdweb/disbandspdweb.go

# final stage
FROM alpine
WORKDIR /app
COPY --from=build-env /src/disbandspdweb /app/
COPY --from=build-env /src/tmpl /app/tmpl
COPY --from=build-env /src/static /app/static
EXPOSE 8080
ENTRYPOINT /app/disbandspdweb

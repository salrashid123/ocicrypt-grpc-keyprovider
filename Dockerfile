FROM golang:1.25.1 as build

WORKDIR /app
ADD . /app
RUN go mod download
RUN GOOS=linux GOARCH=amd64 go build -buildvcs=false  -o /app/server
RUN chown root:root /app/server 

# base-debian-root
FROM gcr.io/distroless/base-debian12

COPY --from=build /app/server  /server
COPY --from=build /app/certs/localhost.crt /certs
COPY --from=build /app/certs/localhost.key /certs
COPY --from=build /app/certs/root-ca.crt /certs

EXPOSE 8080
ENTRYPOINT ["/server"]
CMD ["--useInsecure=true", "--grpcport=:8080"]
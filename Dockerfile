FROM golang:1.23 AS build
WORKDIR /app
COPY . /app
RUN go mod tidy
RUN CGO_ENABLED=0 go build -o /app/shortener /app/...

FROM scratch
COPY --from=build /app/shortener /bin/shortener
EXPOSE 3000
CMD ["/bin/shortener"]


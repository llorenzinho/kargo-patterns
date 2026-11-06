# syntax=docker/dockerfile:1

FROM golang:1.27.3 AS build

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=dev

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w -X github.com/llorenzinho/kargo-pattern/app/internal/version.Version=${VERSION} -X github.com/llorenzinho/kargo-pattern/app/internal/version.Commit=${COMMIT}" \
    -o /out/server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/server /server

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/server"]

# Build the Frontend
FROM node:20-alpine as ui-build-stage
WORKDIR /app
COPY ./ui/package.json package.json
COPY ./ui ./
RUN npm install
RUN npm run build
# Build the Backend
FROM golang:1.24 AS build-stage
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd /app/cmd
COPY internal /app/internal
COPY external /app/external
RUN CGO_ENABLED=0 GOOS=linux go build -o /kasegu ./cmd/kasegu
# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage
WORKDIR /
COPY --from=build-stage /kasegu /kasegu
COPY --from=ui-build-stage /app/dist /static
EXPOSE 1323
CMD ["/kasegu"]
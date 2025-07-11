FROM golang:1.24 AS server-build
WORKDIR /app/backend
COPY . .
RUN go mod download

# compile with static linking
RUN CGO_ENABLED=0 go build -o server 

# -----------------------------------------------------

FROM golang:1.24 AS runtime
WORKDIR /app/backend

COPY --from=server-build /app/backend /app/backend

EXPOSE 8080
ENTRYPOINT [ "/app/backend/server", "--config", ".secrets/conf.yaml" ]
FROM golang:1.21

WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o main .

# Exposer le port dynamiquement
EXPOSE 8080

# Démarrer l'application
CMD ["./main"]
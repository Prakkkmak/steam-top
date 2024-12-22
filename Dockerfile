FROM golang:1.21

WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o main .

# Exposer le port dynamiquement
EXPOSE ${PORT}

# Démarrer l'application
CMD ["./main"]
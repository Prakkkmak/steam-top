# Utiliser une image Go officielle
FROM golang:1.23

# Définir le répertoire de travail
WORKDIR /app

# Copier les fichiers dans le conteneur
COPY . .

# Télécharger les dépendances
RUN go mod tidy

# Compiler l'application
RUN go build -o main .

# Exposer le port 8080
EXPOSE 8080

# Démarrer l'application
CMD ["./main"]
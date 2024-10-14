# Mise en place de l'infrastructure de base pour le Convertisseur JSON-LD

Objectif : Créer la structure de base du projet en Go pour un convertisseur de documents en JSON-LD utilisant le vocabulaire Schema.org.

## Tâches :

1. Créez un nouveau répertoire pour le projet nommé "json-ld-converter".

2. Initialisez un nouveau module Go dans ce répertoire avec la commande :
   ```
   go mod init github.com/votre-username/json-ld-converter
   ```

3. Créez la structure de répertoires suivante dans le projet :
   ```
   json-ld-converter/
   ├── cmd/
   │   ├── cli/
   │   └── server/
   ├── internal/
   │   ├── config/
   │   ├── converter/
   │   ├── logger/
   │   ├── parser/
   │   └── schema/
   ├── pkg/
   ├── test/
   └── docs/
   ```

4. Dans le répertoire racine, créez un fichier README.md avec le contenu suivant :
   ```markdown
   # JSON-LD Document Converter

   Version: 0.3.0 Alpha

   This Go-based software converts various document formats (text, PDF, Markdown, HTML) into a detailed JSON-LD representation using Schema.org vocabulary.

   ## Features (To be implemented)
   - Multi-format input support (text, PDF, Markdown, HTML)
   - Schema.org-based JSON-LD output
   - Handling of large documents (up to 120,000 tokens)
   - Parallel processing and efficient memory management
   - CLI and Server components

   ## Installation
   (To be added)

   ## Usage
   (To be added)

   ## Configuration
   (To be added)

   ## Documentation
   (To be added)

   ## License
   (To be added)
   ```

5. Créez un fichier .gitignore à la racine du projet avec le contenu suivant :
   ```
   # Binaries for programs and plugins
   *.exe
   *.exe~
   *.dll
   *.so
   *.dylib

   # Test binary, built with `go test -c`
   *.test

   # Output of the go coverage tool, specifically when used with LiteIDE
   *.out

   # Dependency directories (remove the comment below to include it)
   # vendor/

   # Go workspace file
   go.work

   # IDE-specific files
   .idea/
   .vscode/

   # OS-specific files
   .DS_Store
   Thumbs.db

   # Log files
   *.log

   # Configuration files with sensitive information
   config.yaml
   ```

6. Initialisez un dépôt Git dans le répertoire du projet :
   ```
   git init
   ```

7. Effectuez un premier commit avec les fichiers créés :
   ```
   git add .
   git commit -m "Initial project structure setup"
   ```

8. Créez un fichier main.go vide dans les répertoires cmd/cli et cmd/server.

9. Dans le répertoire internal/logger, créez un fichier logger.go avec une structure de base pour le système de logging.

10. Dans le répertoire internal/config, créez un fichier config.go avec une structure de base pour la gestion de la configuration.

## Notes importantes :
- Ce projet utilisera Go modules pour la gestion des dépendances.
- La structure du projet suit les bonnes pratiques Go pour la séparation des préoccupations.
- Les fichiers main.go dans cmd/cli et cmd/server seront les points d'entrée pour les composants CLI et serveur respectivement.
- Le répertoire internal contiendra le code spécifique à l'application qui ne doit pas être importé par d'autres projets.
- Le répertoire pkg contiendra le code qui pourrait potentiellement être réutilisé par d'autres projets.
- Assurez-vous d'avoir Go (version 1.16 ou supérieure) installé sur votre système avant de commencer.

Veuillez procéder à la mise en place de cette structure de base et informez-moi une fois que c'est fait pour que nous puissions passer à l'étape suivante du développement.
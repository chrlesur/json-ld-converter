# JSON-LD Converter

## Version 0.3.0 Alpha

JSON-LD Converter est un outil en ligne de commande pour convertir divers formats de documents en JSON-LD en utilisant le vocabulaire Schema.org.

## Fonctionnalités Principales

- Support multi-format d'entrée (texte, PDF, Markdown, HTML)
- Sortie JSON-LD basée sur Schema.org
- Architecture modulaire (composants serveur et client CLI)
- Système de journalisation avancé
- Gestion de la configuration via YAML
- Intégration avec différents LLM (Large Language Models) :
  - Claude (Anthropic)
  - GPT (OpenAI)
  - Ollama (pour les modèles locaux)
  - AI.YOU

## Installation

```bash
go get github.com/chrlesur/json-ld-converter
```

## Utilisation

### Conversion simple

```bash
json-ld-converter -i input.txt -o output.jsonld
```

### Traitement par lots

```bash
json-ld-converter batch -d input_directory -o output_directory
```

### Mode interactif

```bash
json-ld-converter interactive
```

### Gestion de la configuration

Afficher la configuration :
```bash
json-ld-converter config --show
```

Modifier une valeur de configuration :
```bash
json-ld-converter config --set-key conversion.max_tokens --set-value 5000
```

### Options additionnelles

- `-e, --engine` : Moteur LLM à utiliser (claude, openai, ollama, aiyou)
- `-m, --model` : Modèle spécifique à utiliser
- `-o, --output` : Fichier de sortie (par défaut : inputfile.jsonld)
- `-i, --instructions` : Instructions supplémentaires pour le LLM
- `--debug` : Active le mode debug pour des logs détaillés
- `--silent` : Mode silencieux (pas de sortie console)

### Structure du Projet

- `cmd/` : Points d'entrée de l'application
- `internal/` : Packages internes
- `pkg/` : Packages réutilisables
- `data/` : Fichiers de données (ex: schéma Schema.org)

## Configuration

Le fichier de configuration par défaut est `config.yaml`. Vous pouvez spécifier un fichier de configuration différent avec l'option `--config`.

## Contribution

Les contributions sont les bienvenues ! Veuillez consulter le fichier CONTRIBUTING.md pour plus de détails.

## Licence

Ce projet est sous licence GPL3. Voir le fichier LICENSE pour plus de détails.
```
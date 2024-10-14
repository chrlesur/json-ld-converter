# JSON-LD Converter

## Version 0.3.0 Alpha

JSON-LD Converter est un outil en ligne de commande pour convertir divers formats de documents en JSON-LD en utilisant le vocabulaire Schema.org.

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

- `--engine` : Spécifier le moteur LLM à utiliser (claude, gpt, ollama, aiyou)
- `--instructions` : Fournir des instructions supplémentaires au LLM
- `--silent` : Mode silencieux (pas de sortie console)
- `--debug` : Mode debug (journalisation détaillée)

## Configuration

Le fichier de configuration par défaut est `config.yaml`. Vous pouvez spécifier un fichier de configuration différent avec l'option `--config`.

## Contribution

Les contributions sont les bienvenues ! Veuillez consulter le fichier CONTRIBUTING.md pour plus de détails.

## Licence

Ce projet est sous licence GPL3. Voir le fichier LICENSE pour plus de détails.
```
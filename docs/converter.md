# Moteur de Conversion JSON-LD

Ce package implémente un moteur de conversion flexible capable de transformer des segments de document en représentations JSON-LD détaillées, en utilisant le vocabulaire Schema.org et des modèles de langage (LLM) externes.

## Utilisation

Pour utiliser le convertisseur JSON-LD dans votre projet :

1. Importez le package :
   ```go
   import "github.com/chrlesur/json-ld-converter/internal/jsonld"
   ```

2. Créez une nouvelle instance du convertisseur :
   ```go
   converter := jsonld.NewConverter(vocabulary, llmClient, maxTokens, additionalInstructions)
   ```

3. Utilisez la méthode `Convert` pour transformer un segment de document :
   ```go
   jsonLD, err := converter.Convert(documentSegment)
   if err != nil {
       // Gérer l'erreur
   }
   // Utiliser jsonLD...
   ```

## Fonctionnalités principales

- Conversion de segments de document en JSON-LD
- Enrichissement sémantique via LLM
- Gestion des structures imbriquées
- Respect des limites de tokens
- Application d'instructions supplémentaires pour une conversion personnalisée

## Gestion des erreurs

Le convertisseur utilise des types d'erreurs personnalisés pour une gestion précise des problèmes potentiels :

- `ConversionError`: pour les erreurs spécifiques à une étape de la conversion
- `TokenLimitError`: lorsque la limite de tokens est dépassée
- `SchemaOrgError`: pour les erreurs liées au vocabulaire Schema.org

## Exemple

```go
segment := &parser.DocumentSegment{Content: "Ceci est un exemple de contenu."}
jsonLD, err := converter.Convert(segment)
if err != nil {
    log.Fatalf("Erreur de conversion : %v", err)
}
fmt.Printf("JSON-LD généré : %v\n", jsonLD)
```

## Notes importantes

- Assurez-vous que le vocabulaire Schema.org et le client LLM sont correctement initialisés avant de créer le convertisseur.
- La limite de tokens doit être définie en fonction des contraintes du LLM utilisé.
- Les instructions supplémentaires permettent d'affiner le comportement de la conversion, mais leur application peut être ignorée en cas d'erreur pour assurer la continuité du processus.
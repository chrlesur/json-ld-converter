"Implémentez le moteur de conversion JSON-LD pour notre projet JSON-LD-CONVERTER en Go. Ce moteur doit convertir des documents structurés en JSON-LD utilisant le vocabulaire Schema.org. Voici les spécifications détaillées :

1. Créez un package 'converter' dans le répertoire 'internal'.

2. Implémentez une structure 'JSONLDConverter' avec les méthodes suivantes :
   - NewJSONLDConverter() : pour l'initialisation
   - Convert(doc *parser.Document) ([]*jsonld.Document, error) : pour la conversion principale

3. La méthode Convert doit :
   - Mapper les éléments du document d'entrée vers des types et propriétés Schema.org appropriés
   - Créer une structure JSON-LD complète
   - Segmenter la sortie en morceaux de 4000 tokens maximum
   - Utiliser la bibliothèque 'github.com/piprate/json-gold/ld' pour la manipulation JSON-LD

4. Implémentez des fonctions auxiliaires pour :
   - Le mapping document vers Schema.org
   - La segmentation du JSON-LD
   - La gestion des références entre segments

5. Utilisez les packages existants :
   - 'internal/logger' pour la journalisation (ex: logger.Info("message"))
   - 'internal/config' pour la configuration (ex: config.Get().SomeValue)
   - 'pkg/tokenizer' pour le comptage des tokens (ex: tokenizer.CountTokens(text))

6. Assurez-vous que le code est thread-safe pour un traitement parallèle.

7. Gérez les erreurs de manière robuste et utilisez le logging pour les rapporter.

8. Ajoutez des commentaires détaillés expliquant la logique de chaque fonction.

9. Créez des tests unitaires complets dans un fichier 'converter_test.go'.

Voici un exemple de structure de base pour commencer :

```go
package converter

import (
    "github.com/piprate/json-gold/ld"
    "github.com/chrlesur/json-ld-converter/internal/config"
    "github.com/chrlesur/json-ld-converter/internal/logger"
    "github.com/chrlesur/json-ld-converter/internal/parser"
    "github.com/chrlesur/json-ld-converter/pkg/tokenizer"
)

type JSONLDConverter struct {
    // Ajoutez les champs nécessaires
}

func NewJSONLDConverter() *JSONLDConverter {
    // Initialisez et retournez une nouvelle instance
}

func (c *JSONLDConverter) Convert(doc *parser.Document) ([]*ld.RDFDataset, error) {
    // Implémentez la logique de conversion complète ici
    // Cette méthode doit appeler d'autres méthodes pour le mapping, la segmentation, etc.
}

func (c *JSONLDConverter) mapToSchemaOrg(doc *parser.Document) (map[string]interface{}, error) {
    // Implémentez la logique de mapping vers Schema.org
}

func (c *JSONLDConverter) segmentJSONLD(jsonld map[string]interface{}) ([]map[string]interface{}, error) {
    // Implémentez la logique de segmentation
}

// Ajoutez d'autres méthodes nécessaires

```

Assurez-vous que chaque fichier ne dépasse pas 3000 tokens. Si nécessaire, divisez la logique en plusieurs fichiers.

Implémentez toute la logique nécessaire pour que le convertisseur fonctionne de bout en bout, en utilisant les packages mentionnés et en respectant les contraintes de tokens."


Implémentez un système de traitement parallèle pour notre projet JSON-LD-CONVERTER en Go. Ce système doit permettre de traiter efficacement de grands volumes de données en utilisant les goroutines et les canaux de Go. Voici les spécifications détaillées :

1. Créez un package 'parallel' dans le répertoire 'internal'.

2. Implémentez une structure 'ParallelProcessor' avec les méthodes suivantes :
   - NewParallelProcessor(numWorkers int) : pour l'initialisation
   - Process(docs []*parser.Document) ([]*ld.RDFDataset, error) : pour le traitement parallèle

3. La méthode Process doit :
   - Diviser les documents d'entrée en tâches
   - Distribuer ces tâches à un pool de workers (goroutines)
   - Collecter les résultats des workers
   - Gérer les erreurs de manière appropriée

4. Implémentez une structure 'Worker' qui :
   - Traite les documents individuels
   - Utilise le JSONLDConverter existant pour la conversion

5. Utilisez les packages existants :
   - 'internal/logger' pour la journalisation
   - 'internal/config' pour la configuration
   - 'internal/converter' pour la conversion JSON-LD

6. Implémentez un mécanisme de contrôle de flux pour éviter de surcharger la mémoire avec trop de tâches en parallèle.

7. Assurez-vous que le code est thread-safe et gère correctement la concurrence.

8. Ajoutez des métriques de performance, comme le temps de traitement total et le nombre de documents traités par seconde.

9. Gérez les erreurs de manière robuste, en permettant la poursuite du traitement même si certaines tâches échouent.

10. Ajoutez des commentaires détaillés expliquant la logique de chaque fonction.

11. Créez des tests unitaires complets dans un fichier 'parallel_test.go'.

Voici un exemple de structure de base pour commencer :

```go
package parallel

import (
    "sync"
    "github.com/piprate/json-gold/ld"
    "github.com/chrlesur/json-ld-converter/internal/config"
    "github.com/chrlesur/json-ld-converter/internal/logger"
    "github.com/chrlesur/json-ld-converter/internal/parser"
    "github.com/chrlesur/json-ld-converter/internal/converter"
)

type ParallelProcessor struct {
    numWorkers int
    converter  *converter.JSONLDConverter
    // Ajoutez d'autres champs nécessaires
}

func NewParallelProcessor(numWorkers int) *ParallelProcessor {
    // Initialisez et retournez une nouvelle instance
}

func (p *ParallelProcessor) Process(docs []*parser.Document) ([]*ld.RDFDataset, error) {
    // Implémentez la logique de traitement parallèle complète ici
    // Cette méthode doit créer des workers, distribuer les tâches et collecter les résultats
}

type Worker struct {
    converter *converter.JSONLDConverter
    // Ajoutez d'autres champs nécessaires
}

func (w *Worker) processDocument(doc *parser.Document) (*ld.RDFDataset, error) {
    // Implémentez la logique de traitement d'un seul document
}

// Ajoutez d'autres méthodes nécessaires
```

Assurez-vous que chaque fichier ne dépasse pas 3000 tokens. Si nécessaire, divisez la logique en plusieurs fichiers.

Implémentez toute la logique nécessaire pour que le traitement parallèle fonctionne efficacement, en utilisant les packages mentionnés et en respectant les meilleures pratiques de Go pour la concurrence."

Ce prompt devrait fournir une base solide pour implémenter le système de traitement parallèle, en s'intégrant avec les composants existants du projet et en respectant les contraintes spécifiées.
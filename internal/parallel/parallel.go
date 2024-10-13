package parallel

import (
    "sync"
    "time"

    "github.com/piprate/json-gold/ld"
    "github.com/chrlesur/json-ld-converter/internal/config"
    "github.com/chrlesur/json-ld-converter/internal/logger"
    "github.com/chrlesur/json-ld-converter/internal/parser"
    "github.com/chrlesur/json-ld-converter/internal/converter"
)

// ParallelProcessor gère le traitement parallèle des documents
type ParallelProcessor struct {
    numWorkers int
    converter  *converter.JSONLDConverter
    config     *config.Config
}

// NewParallelProcessor crée une nouvelle instance de ParallelProcessor
func NewParallelProcessor(numWorkers int, cfg *config.Config) *ParallelProcessor {
    return &ParallelProcessor{
        numWorkers: numWorkers,
        converter:  converter.NewJSONLDConverter(cfg),
        config:     cfg,
    }
}

// Process traite une liste de documents en parallèle
func (p *ParallelProcessor) Process(docs []*parser.Document) ([]*ld.RDFDataset, error) {
    start := time.Now()
    logger.Info("Démarrage du traitement parallèle")

    results := make([]*ld.RDFDataset, len(docs))
    errors := make([]error, len(docs))

    jobs := make(chan int, len(docs))
    var wg sync.WaitGroup

    // Création des workers
    for w := 0; w < p.numWorkers; w++ {
        wg.Add(1)
        go p.worker(w, jobs, docs, results, errors, &wg)
    }

    // Envoi des tâches aux workers
    for i := range docs {
        jobs <- i
    }
    close(jobs)

    // Attente de la fin du traitement
    wg.Wait()

    // Traitement des erreurs
    for i, err := range errors {
        if err != nil {
            logger.Error("Erreur lors du traitement du document %d: %v", i, err)
        }
    }

    duration := time.Since(start)
    docsPerSecond := float64(len(docs)) / duration.Seconds()
    logger.Info("Traitement terminé en %v. %f documents/seconde", duration, docsPerSecond)

    return results, nil
}

// worker est la fonction exécutée par chaque goroutine worker
func (p *ParallelProcessor) worker(id int, jobs <-chan int, docs []*parser.Document, results []*ld.RDFDataset, errors []error, wg *sync.WaitGroup) {
    defer wg.Done()
    for j := range jobs {
        logger.Debug("Worker %d traite le document %d", id, j)
        result, err := p.converter.Convert(docs[j])
        results[j] = result
        errors[j] = err
    }
}
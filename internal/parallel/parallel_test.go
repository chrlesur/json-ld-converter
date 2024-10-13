package parallel

import (
	"testing"

	"github.com/chrlesur/json-ld-converter/internal/config"
	"github.com/chrlesur/json-ld-converter/internal/parser"
)

func TestParallelProcessor(t *testing.T) {
	// Création d'une configuration de test
	cfg := &config.Config{}

	// Création de documents de test
	docs := []*parser.Document{
		{Content: "Document 1"},
		{Content: "Document 2"},
		{Content: "Document 3"},
	}

	// Création du processeur parallèle
	processor := NewParallelProcessor(2, cfg)

	// Exécution du traitement parallèle
	results, err := processor.Process(docs)

	// Vérification des résultats
	if err != nil {
		t.Errorf("Erreur inattendue : %v", err)
	}
	if len(results) != len(docs) {
		t.Errorf("Nombre de résultats incorrect. Attendu : %d, Obtenu : %d", len(docs), len(results))
	}

	// Ajoutez d'autres vérifications selon vos besoins spécifiques
}

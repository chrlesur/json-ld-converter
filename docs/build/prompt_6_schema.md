# Intégration du vocabulaire Schema.org pour le Convertisseur JSON-LD

Objectif : Intégrer une base de données complète du vocabulaire Schema.org dans le projet, implémenter un système de sélection intelligente des propriétés, et permettre des extensions personnalisées.

## Tâches :

1. Dans le répertoire `internal/schema`, créez un fichier `schema.go` avec le contenu suivant :

```go
package schema

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

type SchemaType struct {
	ID          string            `json:"@id"`
	Label       string            `json:"rdfs:label"`
	Comment     string            `json:"rdfs:comment"`
	Properties  []string          `json:"properties,omitempty"`
	SubClassOf  []string          `json:"subClassOf,omitempty"`
	IsPartOf    string            `json:"isPartOf"`
	Source      string            `json:"source"`
	Enumerations map[string]string `json:"enumerations,omitempty"`
}

type SchemaProperty struct {
	ID           string   `json:"@id"`
	Label        string   `json:"rdfs:label"`
	Comment      string   `json:"rdfs:comment"`
	DomainIncludes []string `json:"domainIncludes,omitempty"`
	RangeIncludes []string `json:"rangeIncludes,omitempty"`
	IsPartOf     string   `json:"isPartOf"`
	Source       string   `json:"source"`
}

type SchemaOrg struct {
	Types      map[string]SchemaType
	Properties map[string]SchemaProperty
}

func LoadSchemaOrg(filePath string) (*SchemaOrg, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading Schema.org file: %w", err)
	}

	var rawSchema map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawSchema); err != nil {
		return nil, fmt.Errorf("error unmarshaling Schema.org data: %w", err)
	}

	schema := &SchemaOrg{
		Types:      make(map[string]SchemaType),
		Properties: make(map[string]SchemaProperty),
	}

	for key, value := range rawSchema {
		if strings.HasPrefix(key, "schema:") {
			var schemaType SchemaType
			if err := json.Unmarshal(value, &schemaType); err == nil {
				schema.Types[key] = schemaType
			} else {
				var schemaProperty SchemaProperty
				if err := json.Unmarshal(value, &schemaProperty); err == nil {
					schema.Properties[key] = schemaProperty
				}
			}
		}
	}

	return schema, nil
}

func (s *SchemaOrg) GetType(typeName string) (SchemaType, bool) {
	t, ok := s.Types["schema:"+typeName]
	return t, ok
}

func (s *SchemaOrg) GetProperty(propertyName string) (SchemaProperty, bool) {
	p, ok := s.Properties["schema:"+propertyName]
	return p, ok
}

func (s *SchemaOrg) SuggestProperties(typeName string, content string) []string {
	schemaType, ok := s.GetType(typeName)
	if !ok {
		return nil
	}

	var suggestedProperties []string
	for _, propName := range schemaType.Properties {
		prop, ok := s.GetProperty(strings.TrimPrefix(propName, "schema:"))
		if ok && strings.Contains(strings.ToLower(content), strings.ToLower(prop.Label)) {
			suggestedProperties = append(suggestedProperties, prop.ID)
		}
	}

	return suggestedProperties
}
```

2. Créez un fichier `schema_test.go` dans le même répertoire pour tester l'intégration Schema.org :

```go
package schema

import (
	"testing"
)

func TestLoadSchemaOrg(t *testing.T) {
	schema, err := LoadSchemaOrg("testdata/schema.json")
	if err != nil {
		t.Fatalf("Failed to load Schema.org: %v", err)
	}

	if len(schema.Types) == 0 {
		t.Error("No types loaded from Schema.org")
	}

	if len(schema.Properties) == 0 {
		t.Error("No properties loaded from Schema.org")
	}

	// Test GetType
	person, ok := schema.GetType("Person")
	if !ok {
		t.Error("Failed to get Person type")
	} else if person.Label != "Person" {
		t.Errorf("Unexpected label for Person: %s", person.Label)
	}

	// Test GetProperty
	name, ok := schema.GetProperty("name")
	if !ok {
		t.Error("Failed to get name property")
	} else if name.Label != "name" {
		t.Errorf("Unexpected label for name property: %s", name.Label)
	}

	// Test SuggestProperties
	suggestedProps := schema.SuggestProperties("Person", "John Doe is 30 years old")
	if len(suggestedProps) == 0 {
		t.Error("No properties suggested for Person")
	}
}
```

3. Téléchargez le fichier Schema.org au format JSON depuis https://schema.org/version/latest/schemaorg-current-https.jsonld et placez-le dans un répertoire `internal/schema/data/`.

4. Modifiez le fichier `internal/config/config.go` pour inclure le chemin vers le fichier Schema.org :

```go
type Config struct {
	// ... autres champs existants ...
	Schema struct {
		FilePath string `yaml:"file_path"`
		Version  string `yaml:"version"`
	} `yaml:"schema"`
}
```

5. Mettez à jour le fichier de configuration `config.yaml` pour inclure le chemin vers le fichier Schema.org :

```yaml
# ... autres configurations existantes ...
schema:
  file_path: "internal/schema/data/schemaorg-current-https.jsonld"
  version: "13.0"
```

6. Dans le fichier principal de votre application (par exemple, `cmd/converter/main.go`), ajoutez le code pour charger et utiliser le vocabulaire Schema.org :

```go
package main

import (
	"log"

	"github.com/chrlesur/json-ld-converter/internal/config"
	"github.com/chrlesur/json-ld-converter/internal/schema"
)

func main() {
	// ... code existant pour charger la configuration ...

	cfg := config.Get()

	schemaOrg, err := schema.LoadSchemaOrg(cfg.Schema.FilePath)
	if err != nil {
		log.Fatalf("Failed to load Schema.org vocabulary: %v", err)
	}

	// Exemple d'utilisation
	person, ok := schemaOrg.GetType("Person")
	if ok {
		log.Printf("Person type: %s", person.Label)
		suggestedProps := schemaOrg.SuggestProperties("Person", "John Doe is 30 years old and works as a software engineer")
		log.Printf("Suggested properties for Person: %v", suggestedProps)
	}

	// ... reste du code ...
}
```

## Utilisation du vocabulaire Schema.org :

Pour utiliser le vocabulaire Schema.org dans d'autres parties du projet, importez le package et utilisez-le comme suit :

```go
import "github.com/chrlesur/json-ld-converter/internal/schema"

func processContent(content string, schemaOrg *schema.SchemaOrg) {
	// Déterminez le type de contenu (par exemple, "Person")
	contentType := determineContentType(content)

	// Obtenez les propriétés suggérées
	suggestedProps := schemaOrg.SuggestProperties(contentType, content)

	// Utilisez les propriétés suggérées pour construire votre JSON-LD
	// ...
}
```

## Extensions personnalisées :

Pour ajouter des extensions personnalisées au vocabulaire Schema.org, vous pouvez créer une fonction qui fusionne votre vocabulaire personnalisé avec le vocabulaire Schema.org standard :

```go
func MergeCustomVocabulary(schemaOrg *SchemaOrg, customTypes map[string]SchemaType, customProperties map[string]SchemaProperty) {
	for key, value := range customTypes {
		schemaOrg.Types[key] = value
	}
	for key, value := range customProperties {
		schemaOrg.Properties[key] = value
	}
}
```

Utilisez cette fonction après avoir chargé le vocabulaire Schema.org standard :

```go
customTypes := map[string]SchemaType{
	"schema:CustomType": SchemaType{
		ID:    "schema:CustomType",
		Label: "Custom Type",
		// ... autres champs ...
	},
}
customProperties := map[string]SchemaProperty{
	"schema:customProperty": SchemaProperty{
		ID:    "schema:customProperty",
		Label: "Custom Property",
		// ... autres champs ...
	},
}

MergeCustomVocabulary(schemaOrg, customTypes, customProperties)
```

## Notes importantes :
- Assurez-vous de télécharger régulièrement la dernière version du fichier Schema.org pour maintenir votre vocabulaire à jour.
- La fonction `SuggestProperties` utilise une méthode simple de correspondance de chaînes. Vous pourriez vouloir implémenter des méthodes plus avancées (comme l'analyse NLP) pour une meilleure précision.
- Pensez à implémenter un système de mise en cache pour améliorer les performances lors de l'utilisation répétée du vocabulaire.
- Testez le système avec divers types de contenus pour vous assurer qu'il fonctionne correctement dans tous les cas.
- Lorsque vous ajoutez des extensions personnalisées, assurez-vous qu'elles ne rentrent pas en conflit avec le vocabulaire standard de Schema.org.

Veuillez implémenter cette intégration du vocabulaire Schema.org et effectuer les tests nécessaires. Une fois terminé, nous pourrons passer à l'étape suivante du développement du convertisseur JSON-LD.
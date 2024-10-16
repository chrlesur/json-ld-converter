package llm

import (
	"fmt"
	"strings"
)

func FormatMapToString(m map[string]string) string {
	var result []string
	for k, v := range m {
		result = append(result, fmt.Sprintf("%s: %s", k, v))
	}
	return strings.Join(result, ", ")
}

func BuildPromptWithContext(content string, context *AnalysisContext) string {
	prompt := `Analysez silencieusement le document fourni et produisez une représentation structurée avec les éléments suivants :

	Entités principales :
	- Liste des entités importantes

	Relations clés :
	- Liste des relations importantes entre les entités

	Résumé :
	Un bref résumé du contenu

	N'utilisez pas de caractère # ou * et ne produisez que du texte simple

	Contexte précédent :
    Entités : %s
    Relations : %s
    Résumé : %s
    Nouveau contenu à analyser : %s`

	return fmt.Sprintf(prompt,
		FormatMapToString(context.PreviousEntities),
		strings.Join(context.PreviousRelations, ", "),
		context.Summary,
		content)
}

func UpdateAnalysisContext(response string, prevContext *AnalysisContext) (*AnalysisContext, error) {
	newContext := &AnalysisContext{
		PreviousEntities:  make(map[string]string),
		PreviousRelations: []string{},
		Summary:           prevContext.Summary,
	}

	lines := strings.Split(response, "\n")
	for _, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) >= 3 {
			newContext.PreviousEntities[parts[0]] = parts[2]
			newContext.PreviousRelations = append(newContext.PreviousRelations, parts[1])
		}
	}

	// Mettre à jour le résumé (ceci est une approche simplifiée)
	newContext.Summary += " " + strings.Join(lines, " ")

	return newContext, nil
}

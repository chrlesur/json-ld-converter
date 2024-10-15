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
	prompt := `Analysez le nouveau contenu fourni en tenant compte du contexte existant. Mettez à jour et complétez la représentation ontologique dans le format structuré suivant :

	{ITEMS}
	[Listez ici tous les items existants, en ajoutant ou modifiant selon le nouveau contenu]

	{PROPERTIES}
	[Listez ici toutes les propriétés existantes, en ajoutant de nouvelles si nécessaire]

	{STATEMENTS}
	[Listez ici toutes les déclarations existantes, en ajoutant de nouvelles basées sur le nouveau contenu]

	{END}

	Règles strictes :

	Conservez les QID et PID existants.
	Pour les nouveaux items ou propriétés, utilisez le prochain numéro disponible.
	Mettez à jour les descriptions et aliases existants si de nouvelles informations sont disponibles.
	Ajoutez de nouvelles déclarations sans supprimer les existantes, sauf en cas de contradiction directe.
	En cas de conflit d'information, privilégiez la source la plus récente.
	Assurez-vous que chaque élément est sur une nouvelle ligne.
	Séparez les champs par des barres verticales |.
	Pour les aliases, utilisez des virgules sans espaces.
	Pour les qualificateurs et références, utilisez le format clé:valeur, séparés par des virgules.
	Effectuez cette mise à jour de manière silencieuse, sans commentaires additionnels.
    
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

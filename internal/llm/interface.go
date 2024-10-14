package llm

import (
    "context"
)

// LLMClient définit l'interface commune pour tous les clients LLM
type LLMClient interface {
    // Translate traduit le contenu donné en utilisant le LLM spécifié
    // 
    // Paramètres :
    // - ctx : le contexte pour la gestion des timeouts et des annulations
    // - content : le contenu à traduire
    // - sourceLang : la langue source du contenu
    // - targetLang : la langue cible pour la traduction
    // - additionalInstructions : instructions supplémentaires pour le LLM
    //
    // Retours :
    // - string : le contenu traduit
    // - error : une erreur si quelque chose s'est mal passé pendant la traduction
    Translate(ctx context.Context, content, sourceLang, targetLang, additionalInstructions string) (string, error)
}
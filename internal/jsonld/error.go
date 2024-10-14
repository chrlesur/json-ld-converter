package jsonld

import (
	"fmt"
)

type ConversionError struct {
	Stage string
	Err   error
}

func (e *ConversionError) Error() string {
	return fmt.Sprintf("erreur de conversion lors de l'étape '%s': %v", e.Stage, e.Err)
}

type TokenLimitError struct {
	Limit int
	Count int
}

func (e *TokenLimitError) Error() string {
	return fmt.Sprintf("limite de tokens dépassée : %d tokens (limite : %d)", e.Count, e.Limit)
}

type SchemaOrgError struct {
	Type string
	Err  error
}

func (e *SchemaOrgError) Error() string {
	return fmt.Sprintf("erreur liée au vocabulaire Schema.org pour le type '%s': %v", e.Type, e.Err)
}

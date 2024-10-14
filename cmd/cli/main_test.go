package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBatchCommand teste la fonctionnalité de traitement par lots
func TestBatchCommand(t *testing.T) {
	// Créer un répertoire temporaire pour les tests
	tempDir, err := ioutil.TempDir("", "batch_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Créer quelques fichiers de test
	testFiles := []string{"test1.txt", "test2.txt", "test3.md"}
	for _, file := range testFiles {
		err := ioutil.WriteFile(filepath.Join(tempDir, file), []byte("Test content"), 0644)
		assert.NoError(t, err)
	}

	// Créer un répertoire de sortie
	outputDir := filepath.Join(tempDir, "output")
	err = os.Mkdir(outputDir, 0755)
	assert.NoError(t, err)

	// Exécuter la commande batch
	cmd := newBatchCmd()
	cmd.SetArgs([]string{"--input-dir", tempDir, "--output-dir", outputDir})
	err = cmd.Execute()
	assert.NoError(t, err)

	// Vérifier que les fichiers de sortie ont été créés
	for _, file := range testFiles {
		outputFile := filepath.Join(outputDir, file+".jsonld")
		_, err := os.Stat(outputFile)
		assert.NoError(t, err, "Output file should exist: %s", outputFile)
	}
}

// TestConfigCommand teste la fonctionnalité de gestion de la configuration
func TestConfigCommand(t *testing.T) {
	// Test d'affichage de la configuration
	cmd := newConfigCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--show"})
	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "server:")
	assert.Contains(t, buf.String(), "conversion:")

	// Test de modification de la configuration
	cmd = newConfigCmd()
	cmd.SetArgs([]string{"--set-key", "conversion.max_tokens", "--set-value", "5000"})
	err = cmd.Execute()
	assert.NoError(t, err)

	// Vérifier que la modification a été appliquée
	cmd = newConfigCmd()
	buf = new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--show"})
	err = cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "max_tokens: 5000")
}

// TestInteractiveCommand teste le mode interactif
func TestInteractiveCommand(t *testing.T) {
	// Simuler une entrée utilisateur
	input := "test.txt\noutput.jsonld\nquit\n"
	in := bytes.NewBufferString(input)

	// Capturer la sortie
	out := new(bytes.Buffer)

	// Créer et exécuter la commande interactive
	cmd := newInteractiveCmd()
	cmd.SetIn(in)
	cmd.SetOut(out)
	err := cmd.Execute()

	assert.NoError(t, err)
	assert.Contains(t, out.String(), "Welcome to interactive mode!")
	assert.Contains(t, out.String(), "Enter input file path")
	assert.Contains(t, out.String(), "Enter output file path")
}

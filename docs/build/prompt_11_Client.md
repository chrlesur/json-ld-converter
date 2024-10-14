Client CLI
Objectif : Développer une interface en ligne de commande (CLI) robuste et conviviale pour le convertisseur de documents en JSON-LD, offrant une large gamme de fonctionnalités et d'options pour répondre aux besoins des utilisateurs.

Contexte : Une version de base du CLI a déjà été implémentée. L'objectif est maintenant d'étendre ses capacités et d'améliorer son utilisation.

Tâches principales :

Extension des fonctionnalités de base a) Implémenter la conversion de fichiers uniques avec des options étendues b) Ajouter le support pour le traitement par lots de plusieurs fichiers c) Développer des commandes pour la gestion de la configuration

Amélioration du contrôle des logs a) Implémenter des options pour contrôler le niveau de log en temps réel b) Ajouter un mode silencieux (--silent) pour désactiver la sortie console des logs c) Développer un mode debug (--debug) pour une journalisation très détaillée

Suivi de la progression a) Implémenter un système de barre de progression pour le traitement de grands documents b) Ajouter des estimations de temps restant pour les longues conversions c) Développer un mode verbose pour afficher des informations détaillées sur chaque étape du processus

Sélection et configuration du LLM a) Ajouter des options pour choisir le LLM à utiliser (Claude, GPT, Ollama, AI.YOU) b) Implémenter des paramètres de configuration spécifiques à chaque LLM c) Permettre la spécification d'instructions supplémentaires pour le LLM avec l'option -i

Gestion des versions de documents a) Implémenter des commandes pour comparer différentes versions d'un même document b) Ajouter des options pour spécifier et gérer les versions de Schema.org utilisées

Mode interactif a) Développer un mode interactif pour des conversions à la volée b) Implémenter un système de prompts pour guider l'utilisateur dans le processus de conversion c) Ajouter des fonctionnalités d'auto-complétion et d'aide contextuelle

Optimisation des performances a) Ajouter des options pour contrôler le traitement parallèle et l'utilisation des ressources b) Implémenter des commandes pour afficher des statistiques de performance

Gestion des erreurs et rapports a) Améliorer l'affichage et le formatage des messages d'erreur dans le CLI b) Implémenter des options pour générer des rapports d'erreur détaillés c) Ajouter des suggestions de résolution pour les erreurs courantes

Intégration avec le système de fichiers a) Implémenter des options pour spécifier des chemins d'entrée/sortie complexes b) Ajouter un support pour les wildcards et les expressions régulières dans la sélection de fichiers c) Développer des commandes pour la gestion des fichiers de sortie (par exemple, écrasement, sauvegarde)

Documentation et aide a) Générer une documentation CLI complète avec des exemples d'utilisation b) Implémenter une commande d'aide détaillée pour chaque sous-commande et option c) Ajouter des messages d'aide contextuels et des suggestions d'utilisation

Livrables attendus :

Code source du client CLI étendu et amélioré
Documentation utilisateur détaillée pour le CLI
Exemples de scripts et de commandes pour des cas d'utilisation courants
Suite de tests pour toutes les nouvelles fonctionnalités du CLI
Contraintes et considérations :

Utiliser la bibliothèque Cobra pour la structure du CLI
Assurer la compatibilité avec les versions précédentes du CLI
Optimiser les performances pour une utilisation fluide, même avec de grands ensembles de données
Suivre les meilleures pratiques de conception CLI (par exemple, respect des conventions POSIX)
Prendre en compte l'internationalisation pour les messages et l'aide du CLI
N'hésitez pas à demander des précisions sur l'une de ces tâches ou à suggérer des améliorations supplémentaires pour le CLI.
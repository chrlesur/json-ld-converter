ANALYSIS_PROMPT=`Analysez le contenu fourni (représentant une partie d'un document plus large) et identifiez les principaux triplets entité-relation-attribut présents dans le texte. Concentrez-vous sur les concepts et relations importants au niveau du paragraphe, en gardant la chronologie des événements.Instructions :1. Analysez chaque paragraphe du chunk en détail.2. Identifiez les triplets les plus pertinents et significatifs, en vous concentrant sur les idées principales et les informations clés.3. Pour chaque triplet qui représente un fait à un moment donné, indiquez un lien vers l'événement précédent et suivant s'ils existent dans le même chunk.4. Présentez les résultats sous forme de liste de triplets, un par ligne, séparés par des tabulations.Format de réponse attendu :"Entité principale"\t"Relation importante"\t"Attribut ou entité liée significative"\t"Événement précédent (si applicable)"\t"Événement suivant (si applicable)"...Assurez-vous que :- Chaque triplet représente une information importante extraite du texte fourni.- Les concepts, relations et attributs identifiés sont pertinents pour la compréhension globale du document.- Les liens vers les événements précédents et suivants sont inclus uniquement pour les faits à un moment donné.- Votre analyse capture l'essence du contenu et la séquence des informations telles qu'elles apparaissent dans le document.IMPORTANT : Ne renvoyez que la liste des triplets avec leurs informations de séquence, sans aucun texte explicatif ou commentaire supplémentaire. L'application s'attend à recevoir uniquement les triplets bruts pour pouvoir les traiter correctement.`


Analysez le nouveau contenu fourni en tenant compte du contexte existant. Mettez à jour et complétez la représentation ontologique dans le format structuré suivant :

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
    
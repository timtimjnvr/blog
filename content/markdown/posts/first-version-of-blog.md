<!-- creation-date: 2026-01-24 -->

# Créer mon propre blog

{{summary}}

## Pourquoi et Comment ?

### L'objectif

Pour pouvoir publier quelque part mes études. Motivier l'analyse de sujets de manière approfondie et me forcer à en produire une restitution claire et compréhensible.

### Pourquoi ne pas utiliser un outil existant ?

Je suis finalement parti sur la création d'un outil from _scratch_ malgré le fait que des miriades de solutions existent déjà : 

* des outils SaS : [Substack](https://substack.com/), [Medium](https://medium.com/) ...
* des générateurs de site statique : [Hugo](https://gohugo.io/), [Jekyll](https://jekyllrb.com/), ...


Pourquoi ce choix ?

* Pour pouvoir contrôler le rendu et l'affichage du site sans être limité par un outil clé en main qui n'offrirait pas cette couche de personnalisation.
* Pour en faire un projet applicatif à part entière et ne pas être limité dans les ajouts de fonctionnalités.
* Pour ne pas avoir à apprendre à utiliser un framework de génération custom.
* Car je trouve l'approche de projet intéressante et même _fun_.

### Présentation du projet

#### Définition

J'ai besoin de pouvoir créer de nouveaux articles sans avoir à réfléchir à la mise en page et à la structuration du site. Le MVP contient les fonctionnalités suivantes :

* Pouvoir ajouter un nouvelle section au site facilement. Ajouter une section c'est juste créer un dossier avec un fichier index.md. Ex: `content/markdown/new_section/index.md`. Cette section doit automatiquement être ajoutée à la _nav bar_ du site.
* Pouvoir lister les pages enfants dans une page d'une section donnée à la demande (i.e pouvoir lister les articles de la sections articles par exemple) et définir des affichages de listes _custom_.
* Pouvoir afficher du code, des schémas, des liens dans le site 
* Vérifier que l'ensemble des liens du site statique (locaux ou externes) sont accessibles.
* Déployer en un clic (Au _merge_ !)

Je veux pouvoir écrire en _Markdown_ et que que ces fichiers soient convertis et intégrés automatiquement en pages HTML de mon site.

#### Le processus de génération du site

Tout se passe dans mon projet blog qui est en 2 parties :

* le contenu du blog dans les dossier `content`, `scripts` et  `styles`
* le générateur go dans le dossier `internal` qui traite le contenu et en compile un site statique dans le dossier `target/build`

Processus de génération (flèche pleine : flux de données principal, flèche en pointillés : entrée ou sortie de fichier) :

```d2
vars: {
  d2-legend: Légende {
    a: Étape {
      shape: step
    }
    b: Fichier {
      shape: document
    }
  }
}

GenerateSite: go run .

MdFiles: Fichiers Markdown {
  shape: document
  style.multiple: true
}

PageGeneration: Pour chaque fichier Mardown

Computations: Listing des sections du site {
  shape: step
}

PageGeneration.Enrich: Enrichissements des Markdown {
  shape: step
}

PageGeneration.Conversion: Transformation du Markdown en HTML brut {
  shape: step
}

PageGeneration.Substitution: Enrichissements et Résolutions HTML {
  shape: step
}

PageGeneration.Projection: Projection du contenu HTML dans une page de template {
  shape: step
}

HTMLFiles: Fichiers HTML {
  shape: document
  style.multiple: true
}

PageGeneration.Validations : Vérification {
  shape: step
}

GenerateSite -> Computations
MdFiles -> PageGeneration.Enrich {
  style: {
    stroke-dash: 3
  }
}
Computations -> PageGeneration.Enrich
PageGeneration.Enrich -> PageGeneration.Conversion
Computations -> PageGeneration.Substitution {
  style: {
    stroke-dash: 3
  }
}
PageGeneration.Conversion -> PageGeneration.Substitution
PageGeneration.Substitution -> PageGeneration.Projection
PageGeneration.Projection -> HTMLFiles {
  style: {
    stroke-dash: 3
  }
}
HTMLFiles -> PageGeneration.Validations {
  style: {
    stroke-dash: 3
  }
}
PageGeneration.Projection -> PageGeneration.Validations
```

#### Les fonctionnalités de cette première version

#### La testabilité

#### L'analytics

## Quoi pour la suite ?

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

J'ai besoin de pouvoir créer de nouveaux articles avec des exemples de code, des schémas sans avoir à réfléchir à la mise en page et à la structuration du site.
Je veux pouvoir écrire en _Markdown_ et que que ces fichiers soient convertis et intégrés automatiquement en pages HTML de mon site.

#### Le processus de génération du site

```d2 scale=0.8
direction: right
Fichiers Markdown -> Convertisseur: lecture
Convertisseur -> Goldmark: conversion
Goldmark -> HTML brut
HTML brut -> Template: injection
Template -> "Page HTML finale"
```

#### Les fonctionnalités de cette première version

#### La testabilité

#### L'analytics

## Quoi pour la suite ?

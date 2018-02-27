# SESIM
__Social Engineering Simulator__

Ce programme est gamebook simplifié, il attend les connexions des 
clients sur le port TCP 53511.

Un client a juste à lancer ``nc @ip_serveur 53511`` pour se connecter 
(@ip_serveur vaut 127.0.0.1 si le serveur est lancé en local). Plusieurs 
clients peuvent se connecter en simultané.

Le client voit apparaitre le texte du noeud "start" et doit entrer un
chiffre parmis ceux proposés ou utiliser une information dont il 
dispose selon la situation. Lorsqu'il est possible d'utiliser une 
information le texte doit l'indiquer.

Il est recommandé aux utilisateurs de prendre en notes les informations 
qui leur paraissent pertinentes pour progresser.

Ce programme peut être réutilisé pour n'importe quel scénario, il suffit 
simplement de créer de nouveaux noeuds au format suivant : 
(Attention en yaml l'indentation est très importante ainsi que les tirets)

````
nodes:
  - node :
      name: un_nom_de_noeud
      text: >
        Un description sur \n
        plusieurs lignes. Chaque \n sera remplacé par un 
        retour chariot. 
      # On peut se passer du engine si l'on ne veut qu'afficher de l'information
      engine:
        - regex: (?i)^une_expression_régulière$
          dest: le_nom_d_un_autre_noeud
        - regex: (?i)^une_autre$
          dest: un_troisieme_noeud

  - node : 
      name: le_nom_d_un_autre_noeud
      text: >
        Desc...

  - node : 
      name: un_troisieme_noeud
      text: >
        Desc...

````

Un point sur les expression régulières, pour ne pas frustrer les 
utilisateurs on préfèrera utiliser le symbole (?i) pour rendre 
l'expression insenssible à la casse.

Les symboles ^ et $ permettent de valider que l'utilisateur n'a 
pas entré d'autres caractères que ceux attendus.

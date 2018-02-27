package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

/* Les deux structures sivante sont initialisées lors du
unMarshal qui parse le fichier text.yaml */

/* Type contenant chaque noeud avec le texte à afficher
et les expressions régulières pour atteindre les prochains noeuds*/
type Node struct {
	Node struct {
		Name   string
		Text   string
		Engine []map[string]string
	}
}

/* Un tableau dynamique (slice) de noeuds */
type text struct {
	Nodes []Node
}

/* Cette structure contient des informations sur l'état interne
et permet le retour au noeud précédent lorsqu'il n'y a pas de regex prévue */
type state struct {
	currentNode string
	lastNode    string
}

/* Cette structure permet de vérfier récuperer les différentes
transitions possibles depuis le noeud courant */
type regdest struct {
	regex string
	dest  string
}

var mytext text

func main() {

	filename, _ := filepath.Abs("./text.yaml")
	textFile, err := ioutil.ReadFile(filename)
	checkErrors(err)

	// On parse le fichier text.yaml
	err = yaml.UnmarshalStrict(textFile, &mytext)
	checkErrors(err)

	// On ouvre en socket en écoute sur le port 53511
	service := ":53511"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkErrors(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkErrors(err)

	// Pour chaque demande de connexion on crée une go routine, associé au client
	for {
		conn, err := listener.Accept()
		if err != nil {
			//LOG
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	// Fermeture de la connexion si la fonction se termine
	defer func() {
		fmt.Println("Connection closed")
		conn.Close()
	}()

	currentState := state{"start", "start"}
	timeout := 20 * time.Minute // Timeout de 20 mins
	bufReader := bufio.NewReader(conn)

	for {
		// Récupération du text et regex/dest du noeud courant
		var regex []regdest
		currentText := (*string)(nil)
		for _, v := range mytext.Nodes {
			if v.Node.Name == currentState.currentNode {
				currentText = &v.Node.Text
				for _, v := range v.Node.Engine {
					regex = append(regex, regdest{v["regex"], v["dest"]})
				}
				break
			}
		}
		// Cas d'erreur avec une transition vers un noeud inconnu
		if currentText == nil {
			error1 := "Erreur : Paragraphe non trouvé"
			currentText = &error1
			currentState = state{currentState.lastNode, "start"}
			// Envoi de l'erreur
			_, err := conn.Write([]byte(*currentText))
			if err != nil {
				fmt.Println(err)
				return
			}
			continue
		}

		/* Formattage du texte pour l'affichage dans un netcat, on remplace les "\n" par des 0x0a*/
		modif := strings.Replace(*currentText, string([]rune{92, 110}), string([]rune{10}), -1)
		// Suppression de l'espace après chaque retour à la ligne
		modif = strings.Replace(modif, string([]rune{10, 32}), string([]rune{10}), -1)
		currentText = &modif
		modifFormat := formatText(*currentText, 70)
		modifFormat = "\x0a" + modifFormat + "> "
		currentText = &modifFormat

		// Ecriture du texte dans la connexion TCP
		_, err := conn.Write([]byte(*currentText))
		if err != nil {
			fmt.Println(err)
			return
		}

		// Definition du timeout
		conn.SetReadDeadline(time.Now().Add(timeout))

		// Lecture de l'entrée utilisateur en TCP jusqu'au premier "\n"
		userInput, err := bufReader.ReadBytes('\n')
		if err != nil {
			//LOG
			fmt.Println(err)
			return
		}
		userInput = userInput[:len(userInput)-1]

		// Analyse de la structure regex si on est sur un noeud intermédiaire
		if len(regex) != 0 {
			for _, v := range regex {
				match, err := regexp.Match(v.regex, userInput)
				if err != nil {
					fmt.Println(err)
					return
				}
				// La première regex qui marche casse la boucle et on passe au noeud suivant
				if match {
					currentState = state{v.dest, currentState.currentNode}
					break
				}
			}
		} else { // Sinon on revient au noeud précédent.
			currentState = state{currentState.lastNode, "start"}
		}
	}
}

/* Permet de formatter le texte avec width caractères par ligne
en insérant des 0x0a dans le texte */
func formatText(text string, width int) string {
	modified := []rune(text)
	wCounter := 0
	offset := 0
	for i, _ := range modified {
		if text[i] != '\x0a' {
			if wCounter < width {
				wCounter++
			} else {
				for offset = i; modified[offset] != '\x20'; offset-- {
					if offset == 0 {
						break
					}
				}
				modified = append(modified[:offset], append([]rune{10}, modified[offset+1:]...)...)
				wCounter = 0 + i - offset
			}
		} else {
			wCounter = 0
		}
	}
	return string(modified)
}

func checkErrors(err error) {
	if err != nil {
		panic(err)
	}
}

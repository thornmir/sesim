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

type Node struct {
	Node struct {
		Name   string `yaml:"name"`
		Text   string `yaml:"text"`
		Engine []map[string]string
	}
}

type text struct {
	Nodes []Node
}

type regdest struct {
	regex string
	dest  string
}

type state struct {
	currentNode string
	lastNode    string
}

var mytext text

func main() {

	filename, _ := filepath.Abs("./text.yaml")
	textFile, err := ioutil.ReadFile(filename)
	checkErrors(err)

	err = yaml.UnmarshalStrict(textFile, &mytext)
	checkErrors(err)

	service := ":53511"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkErrors(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkErrors(err)

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
	// Close connection on function exit
	defer func() {
		fmt.Println("Connection closed")
		conn.Close()
	}()

	currentState := state{"start", "start"}
	timeout := 20 * time.Minute //timeout de 20 mins
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
		if currentText == nil {
			error1 := "Erreur : Paragraphe non trouvé"
			currentText = &error1
			currentState = state{currentState.lastNode, "start"}
			// Send error
			_, err := conn.Write([]byte(*currentText))
			if err != nil {
				fmt.Println(err)
				return
			}
			continue
		}
		modif := strings.Replace(*currentText, string([]rune{92, 110}), string([]rune{10}), -1)
		modif = strings.Replace(modif, string([]rune{10, 32}), string([]rune{10}), -1)
		currentText = &modif
		modifFormat := formatText(*currentText, 70)
		modifFormat = "\x0a" + modifFormat + "> "
		currentText = &modifFormat

		// Ecriture du text dans la connexion TCP
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

		if len(regex) != 0 {
			// Analyse de la structure regex
			for _, v := range regex {
				match, err := regexp.Match(v.regex, userInput)
				if err != nil {
					fmt.Println(err)
					return
				}
				if match {
					currentState = state{v.dest, currentState.currentNode}
					break
				}
			}
		} else {
			currentState = state{currentState.lastNode, "start"}
		}
	}
}

// Insert 0x0a every 'width' characters
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

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Player struct {
	ID    int
	IP    string
	Port  string
	Name  string
	Board Board
}

type Board struct {
	/*
		- 0: case intouchée, non dévoilée
		- 1: coup dans l’eau
		- 2: bateau touché
		- 3: bateau coulé
	*/
	NbBoatsLeft int // nombre de bateaux restants
	Board       [10][10]string
	Boats       [5][4]int // 5 bateaux avec xDebut yDebut xFin yFin
}

func (b *Board) InitBoard() {
	b.NbBoatsLeft = 5
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			b.Board[i][j] = "0"
		}
	}
}

func (b *Board) InitBoat() {
	//horizontal: 0, vertical: 1
	orientation := [2]int{0, 1}

	rand.Seed(time.Now().UnixNano())

	//init boats
	for i := 0; i < 5; i++ {
		//init boat randomly
		b.Boats[i][0] = rand.Intn(10)
		b.Boats[i][1] = rand.Intn(10)

		//check if the boat is horizontal or vertical
		if orientation[rand.Intn(2)] == 0 {
			//horizontal
			b.Boats[i][2] = b.Boats[i][0]
			b.Boats[i][3] = b.Boats[i][1] + rand.Intn(4) + 1
		} else {
			//vertical
			b.Boats[i][2] = b.Boats[i][0] + rand.Intn(4) + 1
			b.Boats[i][3] = b.Boats[i][1]
		}

		//check if the boat is out of the board
		if b.Boats[i][2] > 9 || b.Boats[i][3] > 9 {
			i--
		}

		//check if the boat is on another boat by going through all the boats
		for j := 0; j < i; j++ {
			if (b.Boats[i][0] >= b.Boats[j][0] && b.Boats[i][0] <= b.Boats[j][2] && b.Boats[i][1] >= b.Boats[j][1] && b.Boats[i][1] <= b.Boats[j][3]) ||
				(b.Boats[i][2] >= b.Boats[j][0] && b.Boats[i][2] <= b.Boats[j][2] && b.Boats[i][3] >= b.Boats[j][1] && b.Boats[i][3] <= b.Boats[j][3]) {
				i--
				break
			}
		}
	}

	//display all coordinates of the boats
	for i := 0; i < 5; i++ {
		fmt.Println(b.Boats[i][0], b.Boats[i][1], b.Boats[i][2], b.Boats[i][3])
	}

}

func (b *Board) SendBoard(w http.ResponseWriter, req *http.Request) {
	//send state of the board to the client

	switch req.Method {
	case http.MethodGet:
		for i := 0; i < 10; i++ {
			for j := 0; j < 10; j++ {
				w.Write([]byte(b.Board[i][j]))
			}
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (b *Board) NbBoats(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		w.Write([]byte(strconv.Itoa(b.NbBoatsLeft)))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (b *Board) Hit(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		if err := req.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Something went bad")
			return
		}

		if _, ok := req.PostForm["x"]; !ok {
			fmt.Fprintln(w, "Position x is missing")
			return
		}

		if _, ok := req.PostForm["y"]; !ok {
			fmt.Fprintln(w, "Position y is missing")
			return
		}

		x, err := strconv.Atoi(req.PostForm["x"][0])
		if err != nil {
			fmt.Fprintln(w, "Position x is not a number")
			return
		}

		y, err := strconv.Atoi(req.PostForm["y"][0])
		if err != nil {
			fmt.Fprintln(w, "Position y is not a number")
			return
		}

		if x < 0 || x > 9 || y < 0 || y > 9 {
			fmt.Fprintln(w, "Position out of the board")
			return
		}

		if b.Board[x][y] == "0" {
			//check if there is a boat
			for i := 0; i < len(b.Boats); i++ {
				if x >= b.Boats[i][0] && x <= b.Boats[i][2] && y >= b.Boats[i][1] && y <= b.Boats[i][3] {
					b.Board[x][y] = "2"
					fmt.Fprintln(w, "Hit")

					if b.CheckBoatSunk(x, y) {
						b.NbBoatsLeft--
						fmt.Fprintln(w, "Boat sunk")

						for j := b.Boats[i][0]; j <= b.Boats[i][2]; j++ {
							for k := b.Boats[i][1]; k <= b.Boats[i][3]; k++ {
								b.Board[j][k] = "3"
							}
						}

					}
					return
				}
			}

			b.Board[x][y] = "1"
			fmt.Fprintln(w, "Miss")

		} else if b.Board[x][y] == "1" || b.Board[x][y] == "2" {
			fmt.Fprintln(w, "Already hit")
		} else {
			fmt.Fprintln(w, "Something went bad")
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(w, "Method not allowed")
	}
}

func (b *Board) CheckBoatSunk(x int, y int) bool {
	for i := 0; i < len(b.Boats); i++ {
		if x >= b.Boats[i][0] && x <= b.Boats[i][2] && y >= b.Boats[i][1] && y <= b.Boats[i][3] {
			for x := b.Boats[i][0]; x <= b.Boats[i][2]; x++ {
				for y := b.Boats[i][1]; y <= b.Boats[i][3]; y++ {
					if b.Board[x][y] != "2" {
						return false
					}
				}
			}
			return true
		}
	}
	return false
}

func (b *Board) CheckWin() bool {
	return b.NbBoatsLeft == 0
}

func ConvertToChar(value string) string {
	switch value {
	case "0":
		return "_"
	case "1":
		return "X"
	case "2":
		return "T"
	case "3":
		return "C"
	default:
		return " "
	}
}

func ConvertStringToBoard(text string) [10][10]string {
	var board [10][10]string
	var k int
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			board[i][j] = string(text[k])
			k++
		}
	}
	return board
}

func (p *Player) DisplayBoard() {

	fmt.Print("    ")
	for i := 0; i < 10; i++ {
		fmt.Print(i, "   ")
	}
	fmt.Println()

	for i := 0; i < 10; i++ {
		fmt.Print(i, " | ")
		for j := 0; j < 10; j++ {
			fmt.Print(ConvertToChar(p.Board.Board[i][j]), " | ")
		}
		fmt.Print(i)
		fmt.Println()
	}

	//afficher les numéros de colonne
	fmt.Print("    ")
	for i := 0; i < 10; i++ {
		fmt.Print(i, "   ")
	}
	fmt.Println()

	fmt.Println()
	fmt.Println()
}

func (p *Player) SendName(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, p.Name)
}

func (p *Player) GetName(id int) {
	//make a http request to the server to get the name of the player
	resp, err := http.Get("http://" + p.IP + ":" + p.Port + "/name")
	if err != nil {
		fmt.Println("Error while getting the name of the player")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error while reading the name of the player")
	}

	//set the name and id of the player
	p.Name = string(body)
	p.ID = id
}

func (p *Player) DisplayName() {
	fmt.Printf("- %s (id: %d)\n", strings.Replace(p.Name, "\n", "", -1), p.ID)
}

func (p *Player) GetBoard() {
	//make a http request to the server to get the board of the player
	resp, err := http.Get("http://" + p.IP + ":" + p.Port + "/board")
	if err != nil {
		fmt.Println("Error while getting the board of the player")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error while reading the board of the player")
	}

	//set the board of the player
	p.Board.Board = ConvertStringToBoard(string(body))
}

func (p *Player) GetNbBoats() {
	//make a http request to the server to get the number of boats of the player
	resp, err := http.Get("http://" + p.IP + ":" + p.Port + "/boats")
	if err != nil {
		fmt.Println("Error while getting the number of boats of the player")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error while reading the number of boats of the player")
	}

	//set the number of boats of the player
	p.Board.NbBoatsLeft, _ = strconv.Atoi(string(body))
}

func main() {

	var player Player
	var players []Player
	var addPlayer string
	var idPlayer int

	//start the server
	http.HandleFunc("/board", player.Board.SendBoard)
	http.HandleFunc("/boats", player.Board.NbBoats)
	http.HandleFunc("/hit", player.Board.Hit)
	http.HandleFunc("/name", player.SendName)

	//init the board
	player.Board.InitBoard()

	//add boat to the board
	player.Board.InitBoat()

	//Demander le port sur lequel le serveur va écouter
	fmt.Print("Entrer le port sur lequel le serveur va écouter : ")
	fmt.Scan(&player.Port)

	//demander le nom du joueur
	fmt.Print("Entrer votre nom : (Il sera affiché sur l'écran des autres joueurs) ")
	fmt.Scan(&player.Name)

	//start a goroutine
	go func() {
		log.Fatal(http.ListenAndServe(":"+player.Port, nil))
	}()
	for {

		for addPlayer != "n" {
			var player Player
			fmt.Print("Entrer l'adresse ip du joueur : ")
			fmt.Scan(&player.IP)
			fmt.Print("Entrer le port du joueur : ")
			fmt.Scan(&player.Port)
			players = append(players, player)

			//verifier que le joueur n'est pas dans la liste des joueurs
			for i := 0; i < len(players)-1; i++ {
				if players[i].IP == players[len(players)-1].IP && players[i].Port == players[len(players)-1].Port {
					fmt.Println("Ce joueur est déjà dans la liste des joueurs")
					players = players[:len(players)-1]
				}
			}

			//verifier que le serveur est bien lancé
			resp, err := http.Get("http://" + players[len(players)-1].IP + ":" + players[len(players)-1].Port + "/name")
			if err != nil {
				fmt.Println("Le Joueur n'est pas connecté")
				players = players[:len(players)-1]
			}
			defer resp.Body.Close()

			//demander si on veut ajouter un autre joueur
			fmt.Print("Voulez vous ajouter un autre joueur ? (y/n) ")
			fmt.Scan(&addPlayer)
		}

		//On recupere le nom des joueurs
		for i := 0; i < len(players); i++ {
			players[i].GetName(i)
		}

		//afficher le nom des joueurs
		fmt.Println("Joueurs disponible : ")
		for i := 0; i < len(players); i++ {
			players[i].DisplayName()
		}

		//choisir contre qui jouer
		fmt.Print("Entrer l'id du joueur contre qui vous voulez jouer : ")
		fmt.Scan(&idPlayer)

		//On recupere le nombre de bateaux du joueur
		players[idPlayer].GetNbBoats()
		//On affiche le nombre de bateaux du joueur
		fmt.Println("Il reste", players[idPlayer].Board.NbBoatsLeft, "bateaux à couler ! Sur la grille de", players[idPlayer].Name, "")

		//afficher le board du joueur
		players[idPlayer].GetBoard()
		players[idPlayer].DisplayBoard()

		//Demander les coordonnées de la case à attaquer
		var x, y int
		fmt.Print("Entrer les coordonnées de la case à attaquer : ")
		fmt.Scan(&x, &y)

		//envoyer les coordonnées de la case à attaquer
		resp, err := http.PostForm("http://"+players[idPlayer].IP+":"+players[idPlayer].Port+"/hit", url.Values{"x": {strconv.Itoa(x)}, "y": {strconv.Itoa(y)}})
		if err != nil {
			fmt.Println("Error while sending the hit")
		}

		//recuperer le resultat de l'attaque
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error while reading the hit")
		}
		fmt.Println(string(body))
	}
}

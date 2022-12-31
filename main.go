package main

import (
	"fmt"
	"net/http"
	"strconv"
)

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
	//init boats randomly
	b.Boats[0] = [4]int{0, 0, 0, 1}
	b.Boats[1] = [4]int{1, 0, 1, 2}
	b.Boats[2] = [4]int{2, 0, 2, 3}
	b.Boats[3] = [4]int{3, 0, 3, 4}
	b.Boats[4] = [4]int{4, 0, 4, 5}
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
			fmt.Println("Something went bad")
			fmt.Fprintln(w, "Something went bad")
			return
		}

		for key, value := range req.PostForm {
			fmt.Println(key, "=>", value)
		}

		if _, ok := req.PostForm["x"]; !ok {
			fmt.Println("Position x is missing")
			fmt.Fprintln(w, "Position x is missing")
			return
		}

		if _, ok := req.PostForm["y"]; !ok {
			fmt.Println("Position y is missing")
			fmt.Fprintln(w, "Position y is missing")
			return
		}

		x, err := strconv.Atoi(req.PostForm["x"][0])
		if err != nil {
			fmt.Println("Position x is not a number")
			fmt.Fprintln(w, "Position x is not a number")
			return
		}

		y, err := strconv.Atoi(req.PostForm["y"][0])
		if err != nil {
			fmt.Println("Position y is not a number")
			fmt.Fprintln(w, "Position y is not a number")
			return
		}

		if x < 0 || x > 9 || y < 0 || y > 9 {
			fmt.Println("Position out of the board")
			fmt.Fprintln(w, "Position out of the board")
			return
		}

		if b.Board[x][y] == "0" {
			//check if there is a boat
			for i := 0; i < len(b.Boats); i++ {
				if x >= b.Boats[i][0] && x <= b.Boats[i][2] && y >= b.Boats[i][1] && y <= b.Boats[i][3] {
					b.Board[x][y] = "2"
					fmt.Println("Hit")
					fmt.Fprintln(w, "Hit")

					if b.CheckBoatSunk(x, y) {
						b.NbBoatsLeft--
						fmt.Println("Boat sunk")
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
			fmt.Println("Miss")
			fmt.Fprintln(w, "Miss")

		} else if b.Board[x][y] == "1" || b.Board[x][y] == "2" {
			fmt.Println("Already hit")
			fmt.Fprintln(w, "Already hit")
		} else {
			fmt.Println("Something went bad")
			fmt.Fprintln(w, "Something went bad")
		}

		fmt.Fprintln(w, b.CheckWin())

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Println("Method not allowed")
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
	//check if all boats are sunk
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

func DisplayBoard(text string) {
	var board [10][10]string
	var k int
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			board[i][j] = string(text[k])
			k++
		}
	}

	fmt.Print("    ")
	for i := 0; i < 10; i++ {
		fmt.Print(i, "   ")
	}
	fmt.Println()

	for i := 0; i < 10; i++ {
		fmt.Print(i, " | ")
		for j := 0; j < 10; j++ {
			fmt.Print(ConvertToChar(board[i][j]), " | ")
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

func main() {

	//init the board
	var board Board
	board.InitBoard()

	//add boat to the board
	board.InitBoat()

	var text = "3301000000222000000022221000002222210000222222000000000000000000000000000000000000000000000000000000"
	DisplayBoard(text)

	//start the server
	http.HandleFunc("/board", board.SendBoard)
	http.HandleFunc("/boats", board.NbBoats)
	http.HandleFunc("/hit", board.Hit)

	http.ListenAndServe(":8080", nil)

}

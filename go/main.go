package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Player struct {
	conn *websocket.Conn
	hand string
	mu   sync.Mutex
}

func (p *Player) setHand(hand string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.hand = hand
}

type Game struct {
	players [2]*Player
}

func (g *Game) run() {
	for _, p := range g.players {
		p.conn.WriteMessage(websocket.TextMessage, []byte("対戦相手が見つかりました。じゃんけんの手を送信してください。"))
	}

	var wg sync.WaitGroup
	for i, p := range g.players {
		wg.Add(1)
		go func(playerIndex int, player *Player) {
			defer wg.Done()
			_, message, err := player.conn.ReadMessage()
			if err != nil {
				log.Println(err)
				// Notify the other player
				otherPlayerIndex := (playerIndex + 1) % 2
				g.players[otherPlayerIndex].conn.WriteMessage(websocket.TextMessage, []byte("対戦相手が接続を切断しました。"))
				return
			}
			player.setHand(string(message))
		}(i, p)
	}

	wg.Wait()

	p1 := g.players[0]
	p2 := g.players[1]

	var resultP1, resultP2 string

	if p1.hand == p2.hand {
		resultP1 = "引き分けです"
		resultP2 = "引き分けです"
	} else if isWin(p1.hand, p2.hand) {
		resultP1 = "あなたの勝ちです"
		resultP2 = "あなたの負けです"
	} else {
		resultP1 = "あなたの負けです"
		resultP2 = "あなたの勝ちです"
	}

	p1.conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("相手の手: %s, 結果: %s", p2.hand, resultP1)))
	p2.conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("相手の手: %s, 結果: %s", p1.hand, resultP2)))

	p1.conn.Close()
	p2.conn.Close()
}

func isWin(hand1, hand2 string) bool {
	if hand1 == "ぐー" && hand2 == "ちょき" {
		return true
	}
	if hand1 == "ちょき" && hand2 == "ぱー" {
		return true
	}
	if hand1 == "ぱー" && hand2 == "ぐー" {
		return true
	}
	return false
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var matcher = make(chan *Player, 1)

func rpsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	player := &Player{conn: conn}
	conn.WriteMessage(websocket.TextMessage, []byte("対戦相手を探しています..."))

	select {
	case opponent := <-matcher:
		game := &Game{players: [2]*Player{player, opponent}}
		go game.run()
	case matcher <- player:
		// Player is waiting for an opponent
	}
}

func main() {
	http.HandleFunc("/ws", rpsHandler)
	fmt.Println("じゃんけんサーバーをポート8080で起動します")

	port := "8080"
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"ldv/tinvest/operations"
	"ldv/tinvest/users"
	"log"
	"math/rand"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gookit/ini/v2"
	"github.com/gorilla/websocket"
)

type SumFloat float64

func (sf SumFloat) String() string {
	return fmt.Sprintf("%.2f", float64(sf))
}

type Acc struct {
	Id   string
	Name string
	Sum  SumFloat
}

type HtmlData struct {
	Accs  []Acc
	Total SumFloat
}

type Account struct {
	Id  string
	Sum string
}

type TaskStatus struct {
	Executing bool      `json:"executing"`
	Percent   int       `json:"percent"`
	TotalSum  string    `json:"totalsum"`
	Accounts  []Account `json:"accounts"`
}

func (t *TaskStatus) clear() {
	t.Accounts = (t.Accounts)[:0]
	t.TotalSum = ""
}

var bearer_token string
var accounts users.AccountsData

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *HtmlData) clear() {
	h.Accs = (h.Accs)[:0]
	h.Total = 0.0
}

type VisibilityMessage struct {
	Type  string `json:"type"`
	State string `json:"state"`
}

//****************************************************************

func task(events chan TaskStatus) {

	var pfl operations.Portfolio
	var htmlData HtmlData

	taskStatus := TaskStatus{Executing: true}

	for taskStatus.Percent < 100 {
		taskStatus.Percent += rand.Intn(10) + 1
		if taskStatus.Percent > 100 {
			taskStatus.Percent = 100
			taskStatus.Executing = false
		}

		htmlData.clear()
		taskStatus.clear()

		for _, a := range accounts.Accounts {
			pfl = operations.GetPortfolio(bearer_token, a.Id)
			i := Acc{a.Id, a.Name, SumFloat(pfl.TotalAmountPortfolio.Sum())}
			htmlData.Accs = append(htmlData.Accs, i)
			htmlData.Total += SumFloat(i.Sum)
			j := Account{a.Id, i.Sum.String()}
			taskStatus.Accounts = append(taskStatus.Accounts, j)
		}
		taskStatus.TotalSum = htmlData.Total.String()

		events <- taskStatus
		log.Println("task() tick...")
		time.Sleep(time.Second * 5)
	}
}

//************************************************************************

func receiveMsg(events chan VisibilityMessage) {
	msg := VisibilityMessage{Type: "visibilityMessage"}
	for {
		events <- msg
		log.Println("receiveMsg() tik...")
		time.Sleep(time.Second * 5)
	}
}

//************************************************************************

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error: ", err)
		return
	}
	defer conn.Close()

	events := make(chan TaskStatus, 10)
	visibilityMsg := make(chan VisibilityMessage, 10)

	go task(events)
	go receiveMsg(visibilityMsg)

	for {
		select {
		case status := <-events:
			data, err := json.Marshal(status)
			if err != nil {
				log.Println("JSON marshal error: ", err)
				continue
			}
			err = conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Println("Write message error: ", err)
				return
			}
			log.Println(status)

		case msg := <-visibilityMsg:
			err = conn.ReadJSON(&msg)
			if err != nil {
				log.Println("Error reading msg from client: ", err)
				return
			}
			log.Println(msg)

		case <-r.Context().Done():
			log.Println("Context Done")
			return
		}
	}
}

func main() {
	err := ini.LoadFiles("t-invest.ini")
	if err != nil {
		log.Fatal(err)
	}

	bearer_token = ini.String("Authorization.token")
	accounts = users.GetAccounts(bearer_token)

	// Обработчик для статических файлов (css, js, изображения)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/ws", wsHandler)

	// Обработчик главной страницы
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var pfl operations.Portfolio

		tmplPath := filepath.Join("templates", "layout.html")
		tmpl, err := template.ParseFiles(tmplPath)
		if err != nil {
			http.Error(w, "Ошибка загрузки шаблона", http.StatusInternalServerError)
			return
		}

		var htmlData HtmlData

		for _, a := range accounts.Accounts {
			pfl = operations.GetPortfolio(bearer_token, a.Id)
			i := Acc{a.Id, a.Name, SumFloat(pfl.TotalAmountPortfolio.Sum())}
			htmlData.Accs = append(htmlData.Accs, i)
			htmlData.Total += SumFloat(i.Sum)
		}

		err = tmpl.Execute(w, htmlData)
		if err != nil {
			http.Error(w, "Ошибка рендеринга шаблона", http.StatusInternalServerError)
			return
		}
	})

	http.ListenAndServe(":8901", nil)
}

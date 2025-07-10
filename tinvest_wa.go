package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"ldv/tinvest/operations"
	"ldv/tinvest/users"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
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

type VisibilityMessage struct {
	Type  string `json:"type"`
	State string `json:"state"`
}

var bearer_token string
var accounts users.AccountsData
var visibilityMessage VisibilityMessage

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *HtmlData) clear() {
	h.Accs = (h.Accs)[:0]
	h.Total = 0.0
}

func goid() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

//****************************************************************

func task(events chan TaskStatus) {
	log.Println("task() start...")
	var pfl operations.Portfolio
	var htmlData HtmlData

	taskStatus := TaskStatus{Executing: true}
	visibilityMessage.State = "visible"

	for {

		if visibilityMessage.Type == "clientExit" {
			log.Println("task(): receive meaasge type clientExit. Done.")
			return
		}

		if visibilityMessage.State == "hidden" {
			time.Sleep(time.Second)
			continue
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
		log.Println("task() tick...", goid())
		time.Sleep(time.Second * 5)
	}
}

//************************************************************************

func receiveMsg(conn *websocket.Conn) {
	log.Println("receiveMsg() start...")
	msg := VisibilityMessage{Type: "visibilityMessage"}
	for {
		err := conn.ReadJSON(&msg)
		if err != nil {
			visibilityMessage.Type = "clientExit"
			log.Println("Error reading msg from client: ", err)
			return
		}
		if msg.Type == "visibilityChange" {
			visibilityMessage.State = msg.State
		}
		log.Println(msg)
		//		events <- msg
		log.Println("receiveMsg() tik...")
		time.Sleep(time.Second)
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
	visibilityMessage.Type = "clientStart"

	go task(events)
	go receiveMsg(conn)

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

		case <-r.Context().Done():
			log.Println("Context Done")
			return
		}
	}
}

func main() {
	exepath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	path := path.Dir(exepath)
	fullname := filepath.Join(path, "t-invest.ini")
	err = ini.LoadFiles(fullname)
	if err != nil {
		log.Fatal(err)
	}

	bearer_token = ini.String("Authorization.token")
	accounts = users.GetAccounts(bearer_token)
	visibilityMessage.State = "visible"

	// Обработчик для статических файлов (css, js, изображения)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/ws", wsHandler)

	// Обработчик главной страницы
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var pfl operations.Portfolio

		tmplPath := filepath.Join(path, "templates", "layout.html")
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

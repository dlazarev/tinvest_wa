package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"ldv/tinvest"
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
	_ "modernc.org/sqlite"
)

// https://invest-brands.cdn-tinkoff.ru/<logoName>x<size>.png
// Таким образом, файл логотипа лежит на CDN Tinkoff по пути
// https://invest-brands.cdn-tinkoff.ru/ с добавлением имени файла и размера.

const (
	logoURL   = `https://invest-brands.cdn-tinkoff.ru/`
	imagePath = `images`
)

type Acc struct {
	Id    string
	Name  string
	Total tinvest.SumFloat
}

type HtmlData struct {
	Accs  []Acc
	Total tinvest.SumFloat
}

type AccDetail struct {
	Account    Acc
	DailyYield tinvest.SumFloat
	Pos        operations.Positions
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

var basePath string

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
			i := Acc{a.Id, a.Name, tinvest.SumFloat(pfl.TotalAmountPortfolio.Sum())}
			htmlData.Accs = append(htmlData.Accs, i)
			htmlData.Total += tinvest.SumFloat(i.Total)
			j := Account{a.Id, i.Total.String()}
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

//************************************************************************

func main() {
	exepath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	basePath = path.Dir(exepath)
	fullname := filepath.Join(basePath, "t-invest.ini")
	dbFilename := filepath.Join(basePath, "t-invest.sqlite")

	err = ini.LoadFiles(fullname)
	if err != nil {
		log.Fatal(err)
	}

	err = initDatabase(dbFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	bearer_token = ini.String("Authorization.token")
	accounts = users.GetAccounts(bearer_token)
	visibilityMessage.State = "visible"

	// Обработчик для статических файлов (css, js, изображения)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	fs = http.FileServer(http.Dir("images"))
	http.Handle("/images/", http.StripPrefix("/images/", fs))

	http.HandleFunc("/ws", wsHandler)

	// Обработчик главной страницы
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var pfl operations.Portfolio

		tmplPath := filepath.Join(basePath, "templates", "layout.html")
		tmpl, err := template.ParseFiles(tmplPath)
		if err != nil {
			http.Error(w, "Ошибка загрузки шаблона", http.StatusInternalServerError)
			return
		}

		var htmlData HtmlData

		for _, a := range accounts.Accounts {
			pfl = operations.GetPortfolio(bearer_token, a.Id)
			i := Acc{a.Id, a.Name, tinvest.SumFloat(pfl.TotalAmountPortfolio.Sum())}
			htmlData.Accs = append(htmlData.Accs, i)
			htmlData.Total += tinvest.SumFloat(i.Total)
		}

		err = tmpl.Execute(w, htmlData)
		if err != nil {
			http.Error(w, "Ошибка рендеринга шаблона", http.StatusInternalServerError)
			log.Println(err)
			return
		}
	})

	// Обработчик страницы брокерского счета
	http.HandleFunc("/acc", func(w http.ResponseWriter, r *http.Request) {
		tmplPath := filepath.Join(basePath, "templates", "acc.html")
		tmpl, err := template.ParseFiles(tmplPath)
		if err != nil {
			http.Error(w, "Ошибка загрузки шаблона", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		var accDetail AccDetail
		accId := (r.URL.Query().Get("id"))
		for _, a := range accounts.Accounts {
			if a.Id == accId {
				pfl := operations.GetPortfolio(bearer_token, a.Id)
				accDetail.Account.Id = a.Id
				accDetail.Account.Name = a.Name
				accDetail.Account.Total = tinvest.SumFloat(pfl.TotalAmountPortfolio.Sum())
				accDetail.DailyYield = tinvest.SumFloat(pfl.DailyYield.Sum())
				accDetail.Pos = operations.GetPositions(bearer_token, accId)
				break
			}
		}

		addOperationsBySecurity(bearer_token, &accDetail)
		updateLogo(&accDetail)

		err = tmpl.Execute(w, accDetail)
		if err != nil {
			http.Error(w, "Ошибка рендеринга шаблона", http.StatusInternalServerError)
			return
		}
	})

	http.ListenAndServe(":8901", nil)
}

// Based on https://github.com/zooraze/chi-vue-spa
// and https://github.com/thefrol/go-chi-yandex-cloud-template

package main

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/thefrol/minimal/bucket"
)

var Router = chi.NewRouter()

const editorHTML = "./web/edit.html"

var bct, _ = bucket.WithName("web-dir")

func init() {
	Router.Handle("/view/*", http.StripPrefix("/view", http.HandlerFunc(viewHandler)))
	Router.Get("/", rootHandler)

	Router.Get("/edit/{key}", editHandler)
	Router.Post("/save/{key}", saveHandler)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/index.html")
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	url := strings.TrimLeft(r.URL.Path, "/")
	rr, err := bct.Get(url)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	io.Copy(w, rr)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	buttonText := "Сохранить" // текст, который будет написан на кнопке

	t, err := bct.GetString(key)
	var nf *bucket.KeyNotFound
	if errors.As(err, &nf) {
		println("Создаем новый файл ", key)
		buttonText = "Создать"
	} else if err != nil {
		fmt.Printf("Ошибка получения файла %v из бакета %v, причина: %v\n", key, bct.Name, err)
		return
	}

	// Создаем html из шаблона
	s := struct {
		Text       string
		ButtonText string
	}{
		Text:       t,
		ButtonText: buttonText,
	}

	tt, err := template.ParseFiles(editorHTML)
	if err != nil {
		fmt.Println("Error on creating template: ", err)
		return
	}
	err = tt.Execute(w, s) // отправили клиенту
	if err != nil {
		fmt.Println("error on parsing template: ", err)
		return
	}
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	defer r.Body.Close()

	//Запишем в бакет
	err := bct.Put(r.Body, key)
	if err != nil {
		fmt.Printf("Не удалось загрузить %v в бакет %v по причине: %+v \n", "", bct.Name, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

}

func main() {
	http.ListenAndServe(":8080", Router)
}

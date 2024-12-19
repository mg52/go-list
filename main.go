package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type Page struct {
	Title string
	Body  []string
}

type DB map[string][]string

var Data DB

func ErrWriter(w http.ResponseWriter, statusCode int, err error) {
	var jsonBytes []byte
	jsonBytes, jsonErr := json.Marshal(map[string]interface{}{
		"err": fmt.Sprintf("%v", err),
	})
	if jsonErr != nil {
		jsonBytes = []byte(fmt.Sprintf("err: %v", err))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(jsonBytes)
}

var templates = template.Must(template.ParseFiles("view.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func toString(val interface{}) (string, error) {
	switch v := val.(type) {
	case string:
		return v, nil
	case int:
		return strconv.Itoa(v), nil
	default:
		return "", fmt.Errorf("unsupported type: %T", v)
	}
}

func mainHandler(w http.ResponseWriter, r *http.Request, input []string) {
	fmt.Println("input:", input, "len:", len(input))
	key := input[1]
	body := Data[key]

	switch r.Method {
	case "GET":
		if input[3] == "data" {
			jsonBytes, err := json.Marshal(map[string]interface{}{
				"status":     "success",
				"statusCode": 200,
				"content":    Data[key],
			})
			if err != nil {
				ErrWriter(w, http.StatusInternalServerError, err)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(jsonBytes)
			return
		} else {
			p := Page{
				Title: key,
				Body:  body,
			}
			renderTemplate(w, "view", &p)
		}
	case "POST":
		m := make(map[string]interface{})
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			ErrWriter(w, http.StatusInternalServerError, err)
			return
		}
		content := m["content"]
		if content == nil {
			return
		}
		strContent, err := toString(content)
		if err != nil {
			ErrWriter(w, http.StatusInternalServerError, err)
			return
		}
		trimmed := strings.TrimSpace(strContent)
		if trimmed == "" {
			ErrWriter(w, http.StatusBadRequest, fmt.Errorf("empty content"))
			return
		}
		Data[key] = append(Data[key], trimmed)

		jsonBytes, err := json.Marshal(map[string]interface{}{
			"status":     "success",
			"statusCode": 200,
			"content":    fmt.Sprintf("%s saved.", strContent),
		})
		if err != nil {
			ErrWriter(w, http.StatusInternalServerError, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonBytes)
		return
	default:
		ErrWriter(w, http.StatusInternalServerError, fmt.Errorf("unsupported method"))
		return
	}
}

// var validPath = regexp.MustCompile("^/([a-zA-Z0-9]+)$")
var validPath = regexp.MustCompile("^/([a-zA-Z0-9]+)(/|)(data+|)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, []string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m)
	}
}

func main() {
	Data = make(DB)
	http.HandleFunc("/", makeHandler(mainHandler))
	http.HandleFunc("/api", makeHandler(mainHandler))
	fmt.Println("Running...")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

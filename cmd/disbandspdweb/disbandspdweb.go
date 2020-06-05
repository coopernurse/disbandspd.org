package main

import (
	"encoding/csv"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var fileLock = &sync.Mutex{}

func main() {
	templates, err := template.ParseGlob("tmpl/*.html")
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/alternatives", HtmlPage(templates, "alternatives.html", "Disband SPD - Alternatives to Policing"))
	http.HandleFunc("/petition", func(w http.ResponseWriter, r *http.Request) {
		args := make(map[string]string)
		page := "petition.html"
		title := "Disband SPD - Petition"
		if r.Method == "POST" {
			errmsg := saveSignup(r)
			if errmsg == "" {
				page = "petition_thanks.html"
				title = "Disband SPD - Petition Signed"
			} else {
				args["error"] = errmsg
				copyFormFields(r, args)
			}
		}
		args["title"] = title
		w.Header().Set("content-type", "text/html; charset=utf-8")
		templates.ExecuteTemplate(w, page, args)
	})
	http.Handle("/static/", http.FileServer(http.Dir("./")))
	http.HandleFunc("/", HtmlPage(templates, "index.html", "Disband SPD - Reimagine Public Safety"))

	log.Println("starting server. writing files to:", csvFilename())
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func HtmlPage(templates *template.Template, page string, title string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		args := map[string]string{
			"title": title,
		}
		w.Header().Set("content-type", "text/html; charset=utf-8")
		templates.ExecuteTemplate(w, page, args)
	}
}

func copyFormFields(r *http.Request, args map[string]string) {
	if r.Form != nil {
		for k, _ := range r.Form {
			args[k] = r.Form.Get(k)
		}
	}
}

func csvFilename() string {
	fname := os.Getenv("PETITION_CSV")
	if fname == "" {
		fname = "petition.csv"
	}
	return fname
}

func saveSignup(r *http.Request) string {
	err := r.ParseForm()
	if err != nil {
		log.Println("ERROR: unable to parse form:", err)
		return "ERROR: Unable to parse form"
	}

	name := strings.TrimSpace(r.Form.Get("name"))
	email := strings.TrimSpace(r.Form.Get("email"))

	// intentionally lame validator.. will scrub dupes and invalid email later
	if strings.Contains(email, "@") && strings.Contains(email, ".") {
		fname := csvFilename()

		fileLock.Lock()
		defer fileLock.Unlock()

		f, err := os.OpenFile(fname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println("ERROR: unable to open csv file", fname, err)
			return "ERROR: Unable to open output file"
		}
		defer f.Close()

		w := csv.NewWriter(f)
		w.Write([]string{time.Now().Format(time.RFC3339), name, email})
		w.Flush()
		err = w.Error()
		if err != nil {
			log.Println("ERROR: unable to write to csv", fname, err)
			return "ERROR: Error writing to CSV"
		}
		return ""
	} else {
		return "Please enter a valid email address"
	}
}

package main

import (
	"encoding/csv"
	"encoding/json"
	"html/template"
	"io"
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
	http.HandleFunc("/petition-json", func(w http.ResponseWriter, r *http.Request) {
		resp := make(map[string]interface{})
		errmsg := saveSignup(r)
		if errmsg == "" {
			resp["success"] = true
		} else {
			resp["error"] = errmsg
		}
		enc := json.NewEncoder(w)
		w.Header().Set("content-type", "application/json; charset=utf-8")
		err := enc.Encode(resp)
		if err != nil {
			log.Println("ERROR writing json", err)
		}
	})
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
	http.HandleFunc("/", serveFile("static/boot/index.html", "text/html; charset=utf-8"))

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

func serveFile(fname string, contentType string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open(fname)
		if err != nil {
			msg := "ERROR: unable to open file " + fname
			log.Println(msg, err)
		}
		defer f.Close()
		w.Header().Set("content-type", contentType)
		_, err = io.Copy(w, f)
		if err != nil {
			msg := "ERROR: unable to copy file " + fname
			log.Println(msg, err)
		}
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

func remoteAddr(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
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
		w.Write([]string{time.Now().Format(time.RFC3339), remoteAddr(r), name, email})
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

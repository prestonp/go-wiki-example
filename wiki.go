package main

import (
  "io/ioutil"
  "net/http"
  "html/template"
  "regexp"
  "errors"
)

type Page struct {
  Title string
  Body template.HTML
}

var templates = template.Must(template.ParseFiles("tmpl/view.html", "tmpl/edit.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
var linkRegexp = regexp.MustCompile(`\[[a-zA-Z0-9]+\]`)

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
  m := validPath.FindStringSubmatch(r.URL.Path)
  if m == nil {
    http.NotFound(w, r)
    return "", errors.New("Invalid Page Title")
  }
  return m[2], nil
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
  http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
  p, err := loadPage(title)
  if err != nil {
    http.Redirect(w, r, "/edit/" + title, http.StatusFound)
    return
  }
  p.Body = toHTML(p.Body)
  render(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
  p, err := loadPage(title)
  if err != nil {
    p = &Page{Title: title}
  }
  render(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
  body := r.FormValue("body")
  p := &Page{Title: title, Body: template.HTML(body)}
  err := p.save()
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  http.Redirect(w, r, "/view/" + title, http.StatusFound)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil {
      http.NotFound(w, r)
      return
    }
    fn(w, r, m[2])
  }
}

func render(w http.ResponseWriter, tmpl string, p *Page) {
  err := templates.ExecuteTemplate(w, tmpl + ".html", p)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

func (p *Page) save() error {
  filename := "data/" + p.Title + ".txt"
  return ioutil.WriteFile(filename, []byte(p.Body), 0600)
}

func toHTML(body template.HTML) template.HTML {
  // escape body first
  html := template.HTMLEscapeString(string(body))

  // replace [link] with <a href="/view/link">link</a>
  html = linkRegexp.ReplaceAllStringFunc(html, func(s string) string {
    term := s[1:len(s)-1]
    return "<a href=\"/view/" + term + "\">" + term + "</a>"
  })

  return template.HTML(html)
}

func loadPage(title string) (*Page, error) {
  filename := "data/" + title + ".txt"
  body, err := ioutil.ReadFile(filename)
  if err != nil {
    return nil, err
  }

  return &Page{ Title: title, Body: template.HTML(body) }, nil
}

func main() {
  http.HandleFunc("/", indexHandler)
  http.HandleFunc("/view/", makeHandler(viewHandler))
  http.HandleFunc("/edit/", makeHandler(editHandler))
  http.HandleFunc("/save/", makeHandler(saveHandler))
  http.ListenAndServe(":8080", nil)
}

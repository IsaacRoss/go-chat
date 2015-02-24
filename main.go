// main
package main

import (
	"encoding/json"
	"flag"
	"github.com/isaacross/trace"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/objx"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// templ represents a single template
type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

type Configuration struct {
	Key    string
	Secret string
}

// serveHTTP handles the HTTP request.
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}
	t.templ.Execute(w, data)
}

func GetConfiguration(f *os.File) Configuration {
	decoder := json.NewDecoder(f)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Fatal("Can not read config")
	}
	return configuration
}

func main() {
	file, _ := os.Open("../conf.json")
	configuration := GetConfiguration(file)
	var addr = flag.String("addr", ":8080", "the address of the app")
	flag.Parse()
	//set up gomniauth
	gomniauth.SetSecurityKey("ThereOnceWasAGirlFromAlsace")
	gomniauth.WithProviders(github.New(configuration.Key, configuration.Secret, "http://localhost:8080/auth/callback/github"))
	r := newRoom()
	r.tracer = trace.New(os.Stdout)
	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/auth/", loginHandler)
	http.Handle("/room", r)
	// start room
	go r.run()
	// start the web server
	log.Println("Starting web server on: ", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

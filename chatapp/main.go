package main

import (
	"flag"
	"fmt"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/facebook"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"text/template"
)

//  this struct type that is responsible for loading, compiling, and delivering our template.
//  compile the template once (using the sync.Once type),
//  keep the reference to the compiled template, and then respond to HTTP requests.
type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

// The sync.Once type guarantees that the function we pass as an argument will only be executed once,
// regardless of how many goroutines are calling ServeHTTP.
// COMMAND+N implement method Handler
// ServeHTTP handles the HTTP request.
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})

	// pass the request r as the data argument to the Execute method:
	// This tells the template to render itself using data that can be extracted from http.Request,
	// which happens to include the host address that we need from the flag
	// so we can access r.Host from Html
	// To use the Host value of http.Request,
	// we can then make use of the special template syntax that allows us to inject data.
	// Update the line where we create our socket in the chat.html file:
	//t.templ.Execute(w, r)

	// instead directly passing request we change to this
	// so we can display user
	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}
	t.templ.Execute(w, data)
}

func main() {

	// by using this flag we can run at any port we want
	// go build -o chat
	//./chat -addr=":3000"
	//Valid hosts aren't just port numbers; you can also specify the IP addresses or
	// other hostnames provided they are allowed in your environment,
	// for example, -addr="192.168.0.1:3000"

	// using injection from build parameters for our Host
	// The call to flag.String returns a type of *string,
	// which is to say it returns the address of a string variable where the value of the flag is stored.
	// To get the value itself (and not the address of the value), we must use the pointer indirection operator, *.
	var addr = flag.String("addr", ":8080", "The addr host of the application.")
	// We must call flag.Parse() that parses the arguments and extracts the appropriate information.
	// Then, we can reference the value of the host flag by using *addr
	flag.Parse()

	// Oauth2
	//Gomniauth requires the SetSecurityKey call because it sends state data between the client and server along with a signature checksum, which ensures that the state values are not tempered with while being transmitted. The security key is used when creating the hash in a way that it is almost impossible to recreate the same hash without knowing the exact security key. You should replace some long key with a security hash or phrase of your choice.
	gomniauth.SetSecurityKey("8928520ef05111eb9a030242ac130003") //uuid
	gomniauth.WithProviders(
		facebook.New("key", "secret",
			"http://localhost:8080/auth/callback/facebook"),
		github.New("key", "secret",
			"http://localhost:8080/auth/callback/github"),
		google.New("14378472304-oaeg1d6hs32nsk6h1av23mb9hopgsldj.apps.googleusercontent.com", "MucNovJBH6e5sdMfBC9myteU",
			"http://localhost:8080/auth/callback/google"),
	)

	// controller
	chatHtmlTemplate := templateHandler{filename: "chat.html"}

	// We use decorator pattern to wrap this chatHtmlTemplate Handler to authHandle.
	// the ServeHTTP on authHandle will get executed first
	http.Handle("/chat", MustAuth(&chatHtmlTemplate))

	// login controller
	http.Handle("/login", &templateHandler{filename: "login.html"})

	// in go if you end with / it become a prefix to accept, example
	// --> /auth/login/google
	// --> /auth/login/facebook
	// --> /auth/callback/google
	// --> /auth/callback/facebook
	// our loginHandler is not and object, it is a function but it statisfied with the handler ServeHTTP method.
	// go allowed this. why we do this because we don't need it to store any state.
	http.HandleFunc("/auth/", loginHandler)

	// The preceding handler function uses http.SetCookie to update the cookie setting MaxAge to -1,
	// which indicates that it should be deleted immediately by the browser.
	// Not all browsers are forced to delete the cookie,
	// which is why we also provide a new Value setting of an empty string,
	// thus removing the user data that would previously have been stored.
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:   "auth",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
		w.Header().Set("Location", "/chat")
		w.WriteHeader(http.StatusTemporaryRedirect)
	})

	// room controller
	// We didn't have to create an instance of AuthAvatar, so no memory was allocated. In our case,
	// this doesn't result in great saving (since we only have one room for our entire application),
	// but imagine the size of the potential savings if our application has thousands of rooms.
	// The way we named the UseAuthAvatar variable means that the preceding code is very easy
	// to read and it also makes our intention obvious.

	//r := newRoom(UseAuthAvatar)

	// Or we want to use Gravatar instead
	r := newRoom(UseGravatar)

	// if user dont want to use any tracer
	//r.trace = trace.New(os.Stdout)

	http.Handle("/room", r)

	// get the room going
	// room in a separate Go routine (notice the go keyword again)
	// so that the chatting operations occur in the background, allowing our main thread to run the web server
	go r.run()

	// start the web server
	log.Println("Starting web server on", *addr)

	// bootstrap function
	// start the web server
	//if err := http.ListenAndServe(":8080", nil); err != nil {
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe", err)
	}

	fmt.Println("Server started...")
}

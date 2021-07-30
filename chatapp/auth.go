package main

import (
	"fmt"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/objx"
	"net/http"
	"strings"
)

type authHandler struct {
	next http.Handler
}

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("auth")

	if err == http.ErrNoCookie || cookie.Value == "" {
		// not authenticated
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}

	if err != nil {
		// some other error
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// success - call the next handler
	h.next.ServeHTTP(w, r)
}

// is is decorator patter, we wrap the handler object anc chain it to next.
// called by main.go on route / to execute this ServeHTTP and then if success return to the ServeHTTP on the wrapped object
func MustAuth(handler http.Handler) http.Handler {
	return &authHandler{next: handler}
}

// loginHandler handles the third-party login process.
// format: /auth/{action}/{provider}
// We do two main things here. First, we use the gomniauth.Provider function to get the provider object
// that matches the object
// specified in the URL (such as google or github).
// Then, we use the GetBeginAuthURL method to get the location where
// we must send users to in order to start the authorization process.
func loginHandler(w http.ResponseWriter, r *http.Request) {
	segs := strings.Split(r.URL.Path, "/")

	fmt.Println(len(segs))
	if len(segs) < 4 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "Sorry the page your looking for is not found")
		return
	}

	action := segs[2]
	provider := segs[3]
	switch action {

	case "login":

		provider, err := gomniauth.Provider(provider)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error when trying to get provider %s: %s", provider, err),
				http.StatusBadRequest)
			return
		}

		loginUrl, err := provider.GetBeginAuthURL(nil, nil)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error when trying to GetBeginAuthURL for %s:%s", provider, err),
				http.StatusInternalServerError)
			return
		}

		w.Header().Set("Location", loginUrl)
		w.WriteHeader(http.StatusTemporaryRedirect)

	case "callback":

		provider, err := gomniauth.Provider(provider)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error when trying to complete auth for %s: %s", provider, err),
				http.StatusInternalServerError)
			return
		}

		cred, err := provider.CompleteAuth(objx.MustFromURLQuery(r.URL.RawQuery))
		if err != nil {
			http.Error(w, fmt.Sprintf("Error when trying to complete auth for %s: %s", provider, err),
				http.StatusInternalServerError)
			return
		}

		usr, err := provider.GetUser(cred)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error when trying to get user from %s: %s",
				provider, err), http.StatusInternalServerError)
			return
		}

		fmt.Println("usr", usr)

		fmt.Println("avatar", usr.AvatarURL())

		authCookieValue := objx.New(map[string]interface{}{
			"name":       usr.Name(),
			"avatar_url": usr.AvatarURL(),
			"email":      usr.Email(),
		}).MustBase64()

		http.SetCookie(w, &http.Cookie{
			Name:  "auth",
			Value: authCookieValue,
			Path:  "/",
		})

		w.Header().Set("Location", "/chat")
		w.WriteHeader(http.StatusTemporaryRedirect)

	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Auth action %s not supported", action)
	}
}

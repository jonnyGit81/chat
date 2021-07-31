package main

import (
	"crypto/md5"
	"fmt"
	"github.com/stretchr/gomniauth"
	gomniauthcommon "github.com/stretchr/gomniauth/common"
	"github.com/stretchr/objx"
	"io"
	"log"
	"net/http"
	"strings"
)

type ChatUser interface {
	UniqueID() string
	AvatarURL() string
}

// It also makes use of a very interesting feature in Go: type embedding.
// We actually embedded the gomniauth/common.User interface type,
// which means that our struct interface implements the interface automatically.
type chatUser struct {
	gomniauthcommon.User
	uniqueID string
}

func (u chatUser) UniqueID() string {
	return u.uniqueID
}

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

		/* Change to chat user
		usr, err := provider.GetUser(cred)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error when trying to get user from %s: %s",
				provider, err), http.StatusInternalServerError)
			return
		}
		*/

		// So now is dynamic implementation of get user gravatar
		usr, err := provider.GetUser(cred)
		if err != nil {
			log.Fatalln("Error when trying to get user from", provider, "-", err)
		}
		chatUser := &chatUser{User: usr}
		m := md5.New()
		io.WriteString(m, strings.ToLower(usr.Email()))
		chatUser.uniqueID = fmt.Sprintf("%x", m.Sum(nil))
		avatarURL, err := avatars.GetAvatarURL(chatUser)
		if err != nil {
			log.Fatalln("Error when trying to GetAvatarURL", "-", err)
		}

		fmt.Println("usr", usr)

		// Here, we have hashed the e-mail address and stored the resulting value in the userid field at the point at which the user logs in.
		// From now on, we can use this value in our Gravatar code instead of hashing the e-mail address for every message.

		/* we chance it to cache in cookie, get from chatuser
		m := md5.New()
		io.WriteString(m, strings.ToLower(usr.Email()))
		userid := fmt.Sprintf("%x", m.Sum(nil))
		*/
		authCookieValue := objx.New(map[string]interface{}{
			//"userid": userid,
			"userid": chatUser.uniqueID,
			"name":   usr.Name(),
			//"avatar_url": usr.AvatarURL(),
			"avatar_url": avatarURL,
			// "email":      usr.Email(), no longer need
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

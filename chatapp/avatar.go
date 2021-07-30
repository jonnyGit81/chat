package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"strings"
)

// ErrNoAvatar is the error that is returned when the
// Avatar instance is unable to provide an avatar URL.
var ErrNoAvatarURL = errors.New("chat: Unable to get an avatar  URL.")

// Avatar represents types capable of representing
// user profile pictures.

type Avatar interface {
	// GetAvatarURL gets the avatar URL for the specified client,
	// or returns an error if something goes wrong.
	// ErrNoAvatarURL is returned if the object is unable to get
	// a URL for the specified client.
	GetAvatarURL(c *client) (string, error)
}

type AuthAvatar struct{}

//create a handy variable called UseAuthAvatar that has the AuthAvatar type but which remains of nil value.
// We can later assign the UseAuthAvatar variable to any field looking for an Avatar interface type.
var UseAuthAvatar AuthAvatar

//Normally, the receiver of a method (the type defined in parentheses before the name)
// will be assigned to a variable so that it can be accessed in the body of the method.
// Since, in our case, we assume the object can have nil value, we can omit a variable
// name to tell Go to throw away the reference.
// This serves as an added reminder to ourselves that we should avoid using it.
func (AuthAvatar) GetAvatarURL(c *client) (string, error) {
	//if url, ok := c.userData["avatar_url"]; ok {
	//	if urlStr, ok := url.(string); ok {
	//		return urlStr, nil
	//	}
	//}
	//return "", ErrNoAvatarURL

	url, ok := c.userData["avatar_url"]
	if !ok {
		return "", ErrNoAvatarURL
	}

	urlStr, ok := url.(string)
	if !ok {
		return "", ErrNoAvatarURL
	}

	return urlStr, nil
}

// We used the same pattern as we did for AuthAvatar:
// we have an empty struct, a helpful UseGravatar variable, and the GetAvatarURL method implementation itself.
// In this method, we follow Gravatar's guidelines to generate an MD5 hash from the e-mail address
// (after we ensured it was lowercase) and append it to the hardcoded base URL using fmt.Sprintf.
type GravatarAvatar struct{}

var UseGravatar GravatarAvatar

func (GravatarAvatar) GetAvatarURL(c *client) (string, error) {
	email, ok := c.userData["email"]
	if !ok {
		return "", ErrNoAvatarURL
	}

	emailStr, ok := email.(string)
	if !ok {
		return "", ErrNoAvatarURL
	}

	m := md5.New()
	io.WriteString(m, strings.ToLower(emailStr))
	return fmt.Sprintf("//www.gravatar.com/avatar/%x", m.Sum(nil)), nil
}

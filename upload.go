package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"path"

	"github.com/stretchr/objx"
)

func uploaderHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.FormValue("userID")
	file, header, err := r.FormFile("avatarFile")
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}

	filename := path.Join("avatars", userID+path.Ext(header.Filename))
	err = ioutil.WriteFile(filename, data, 0777)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}

	authCookie, err := r.Cookie("auth")
	if err != nil {
		log.Fatalln(err)
	}

	authCookieMap := objx.MustFromBase64(authCookie.Value)
	authCookieMap["avatar_url"] = filename
	authCookieResult := objx.New(map[string]interface{}{
		"user_id":    authCookieMap["user_id"],
		"name":       authCookieMap["name"],
		"avatar_url": authCookieMap["avatar_url"],
	}).MustBase64()

	http.SetCookie(w, &http.Cookie{
		Name:  "auth",
		Value: authCookieResult,
		Path:  "/",
	})
	w.Header()["Location"] = []string{"/chat"}
	w.WriteHeader(http.StatusTemporaryRedirect)
}

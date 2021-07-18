package app

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Picture       string `json:"picture"`
}

var googleOauthConfig = oauth2.Config{
	RedirectURL:  "http://localhost:3000/auth/google/callback",
	ClientID:     "286752575099-uafrkf8h2diclh6f5g4p1f4qet2o60d1.apps.googleusercontent.com",
	ClientSecret: "b0AJBFdcEnTw0FApu9YIiq9j",
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
	Endpoint:     google.Endpoint,
}

func googleLoginHandler(w http.ResponseWriter, r *http.Request) {
	// CSRF 공격을 막기위한 임시 키를 만든다
	state := generateStateOauthCookie(&w)
	url := googleOauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func generateStateOauthCookie(w *http.ResponseWriter) string {
	// 만료시간은 24시간
	expiration := time.Now().Add(1 * 24 * time.Hour)
	// 임의의 키 생성을 위하여 16바이트의 변수를 만든다.
	b := make([]byte, 16)
	// 16바이트 변수에 랜덤하게 값을 채운다.
	rand.Read(b)
	// 쿠키로 담기 위하여 URLEncoding 한다.
	state := base64.URLEncoding.EncodeToString(b)
	cookie := &http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(*w, cookie)
	return state
}

func googleAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	// 쿠키의 state
	oauthstate, _ := r.Cookie("oauthstate")
	// google에서 보내온 state값과 쿠키의 state값이 같은지 체크한다.
	if r.FormValue("state") != oauthstate.Value {
		errMsg := fmt.Sprintf("invalid google oauth state cookie:%s, state:%s\n", oauthstate.Value, r.FormValue("state"))
		log.Printf(errMsg)
		// 해커의 공격시도가 있을경우 최소한의 정보 제공을 위하여 root로 리다이렉트 한다.
		http.Error(w, errMsg, http.StatusInternalServerError)
		// http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	data, err := getGoogleUserInfo(r.FormValue("code"))
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var userInfo GoogleUserInfo
	err = json.Unmarshal(data, &userInfo)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 아이디 정보를 쿠키에 저장한다.
	session, err := store.Get(r, "session")
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Values["id"] = userInfo.ID
	session.Values["picture"] = userInfo.Picture
	err = session.Save(r, w)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)

	// fmt.Fprint(w, string(data))
}

var oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

// 구글ㅇ에서 받은 code로 사용자 정보를 가져온다.
func getGoogleUserInfo(code string) ([]byte, error) {
	// 구글에서 받은 코드를 가지고 Token을 받아온다.
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("Failed to Exchange %s\n", err.Error())
	}

	resp, err := http.Get(oauthGoogleUrlAPI + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("Failed to Get UserInfo %s\n", err.Error())
	}

	return ioutil.ReadAll(resp.Body)
}

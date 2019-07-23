package database

import (
    "encoding/base64"
    "github.com/gin-gonic/gin"
    "net/http"
)

const AuthUserKey = "user"

type Accounts map[string]string

type authPair struct {
    value string
    user  string
}

type authPairs []authPair

func (a authPairs) searchCredential(authValue string) (string, bool) {
    if authValue == "" {
        return "", false
    }
    for _, pair := range a {
        if pair.value == authValue {
            return pair.user, true
        }
    }
    return "", false
}

func BasicAuth(accounts Accounts) gin.HandlerFunc {
    pairs := processAccounts(accounts)
    return func(c *gin.Context) {
        user, found := pairs.searchCredential(c.Request.Header.Get(
            "Authorization"))
        if !found {
            c.AbortWithStatus(http.StatusUnauthorized)
            return
        }
        c.Set(AuthUserKey, user)
    }
}

func processAccounts(accounts Accounts) authPairs {
    assert1(len(accounts) > 0, "Empty list of authorized credentials")
    pairs := make(authPairs, 0, len(accounts))
    for user, password := range accounts {
        assert1(user != "", "User can not be empty")
        value := authorizationHeader(user, password)
        pairs = append(pairs, authPair{
            value: value,
            user:  user,
        })
    }
    return pairs
}

func assert1(guard bool, text string) {
    if !guard {
        panic(text)
    }
}

func authorizationHeader(user, password string) string {
    base := user + ":" + password
    return "Basic " + base64.StdEncoding.EncodeToString([]byte(base))
}


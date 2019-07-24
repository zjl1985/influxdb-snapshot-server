package database

import (
    "encoding/base64"
    "fastdb-server/models"
    "fastdb-server/service"
    "github.com/BurntSushi/toml"
    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
    "net/http"
    "strings"
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
            if strings.Compare(c.Request.RequestURI, "/api/menu") == 0 {
                c.AbortWithStatusJSON(http.StatusOK, gin.H{})
            }
            if strings.Compare(c.Request.RequestURI, "/api/getsysinfo") == 0 {
                c.AbortWithStatusJSON(http.StatusOK, gin.H{})
            }
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

func Login(c *gin.Context) {
    token := c.Request.Header.Get("Authorization")
    myToken := authorizationHeader(service.MyConfig.FastUser, service.MyConfig.FastPwd)
    if strings.Compare(token, myToken) == 0 {
        c.JSON(http.StatusOK, gin.H{
            "success": true,
        })
    } else {
        c.JSON(http.StatusOK, gin.H{
            "success": false,
        })
    }
}

func GetMenu(c *gin.Context) {
    _, exists := c.Get(AuthUserKey)
    if !exists {
        c.AbortWithStatusJSON(http.StatusOK, gin.H{})
    }
    var menus models.MenuInfo
    if _, err := toml.DecodeFile("menu.toml", &menus); err != nil {
        logrus.Error(err)
        c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
            "message": err,
        })
    }

    c.JSON(http.StatusOK, menus)
}

func GetSysInfo(c *gin.Context) {
    _, exists := c.Get(AuthUserKey)
    if !exists {
        c.AbortWithStatusJSON(http.StatusOK, gin.H{})
    }
    var info models.Info
    if _, err := toml.DecodeFile("info.toml", &info); err != nil {
        logrus.Error(err)
        c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
            "message": err,
        })
    }
    c.JSON(http.StatusOK, info)
}

func LogOut(c *gin.Context) {
    c.String(http.StatusOK, "ok")
}

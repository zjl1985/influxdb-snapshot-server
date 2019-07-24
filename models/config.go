package models

type Config struct {
    EnableAuth bool
    Delay         int
    Port          string
    Mode          string
    FastDBPort    string
    FastDBIP      string
    RedisPwd      string
    DBPath        string
    FastDBAddress string
    FastUser      string
    FastPwd       string
    WebPath       string
}

type Info struct {
    Title    string `toml:"title" json:"title"`
    Desc     string `toml:"desc" json:"desc"`
    Company  string `toml:"company" json:"company"`
    UserName string `toml:"userName" json:"userName"`
    Email    string `toml:"email" json:"email"`
}

type MenuInfo struct {
    Menus []menu `toml:"menu" json:"menus"`
}

type icon struct {
    Type  string `toml:"type" json:"type"`
    Value string `toml:"value" json:"value"`
    Theme string `toml:"theme" json:"theme"`
}

type menu struct {
    Text     string `toml:"text" json:"text"`
    I18N     string `toml:"i18n" json:"i18n"`
    Link     string `toml:"link" json:"link"`
    Shortcut bool   `toml:"shortcut" json:"shortcut"`
    Icon     icon   `toml:"icon" json:"icon"`
}

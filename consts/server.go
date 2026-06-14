package consts

import "github.com/dishan1223/mutt/internal/config"

var PORT = ":" + config.MustGetEnv("PORT")

const HASH_COST = 10

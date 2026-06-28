package consts

import "github.com/dishan1223/mutt/internal/config"

func GetPort() string {
	return ":" + config.MustGetEnv("PORT")
}

var HASH_COST = config.MustGetEnv("HASH_COST")

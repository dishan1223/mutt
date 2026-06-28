package consts

import "github.com/dishan1223/mutt/internal/config"

var (
	API_KEY_BYTES   = config.MustGetEnv("API_KEY_BYTES")
	MAX_LOG_SIZE    = config.MustGetEnv("MAX_LOG_SIZE")
	MAX_STACK_TRACE = config.MustGetEnv("MAX_STACK_TRACE")
)

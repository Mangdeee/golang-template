package main

import (
	"github.com/felixlambertv/go-cleanplate/config"
	"github.com/felixlambertv/go-cleanplate/internal/app"
)

func main() {
	cfg := config.GetInstance()
	app.Run(cfg)
}

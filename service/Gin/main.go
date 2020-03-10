package main

import (
	"FileCloud/config"
	"FileCloud/route"
)

func main() {
	// gin框架
	router := route.Router()
	router.Run(config.UploadServiceHost)
}

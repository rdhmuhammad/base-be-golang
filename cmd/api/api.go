package main

import (
	"base-be-golang/shared/api"
	"base-be-golang/shared/base"
	"flag"
	"log"

	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rdhmuhammad/base-be-golang/iam-module/shared/adapter/controller"
	"gorm.io/gorm"
)

func main() {
	var envFile string
	flag.StringVar(&envFile, "env", ".env.stag", "Provide env file path")
	flag.Parse()
	err := godotenv.Load(envFile)
	if err != nil {
		log.Println(err)
		panic(err)

	}

	start := api.Default()

	// ========================= REGISTER CONTROLLER =========================
	// IAM MODULE
	start.Register(func(dbConn *gorm.DB, port base.Port, ctrl base.BaseController) api.Router {
		return controller.NewAuthController(dbConn, port, ctrl)
	})
	start.Register(func(dbConn *gorm.DB, port base.Port, ctrl base.BaseController) api.Router {
		return controller.NewUserManagementController(dbConn, port, ctrl)
	})

	// BUSINESS MODULE

	err = start.Start()
	if err != nil {
		panic(err)
	}

}

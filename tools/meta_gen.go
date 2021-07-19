package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path"

	"github.com/coomp/ccs/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {

	u, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Home dir:", u.HomeDir)

	os.MkdirAll(path.Join(u.HomeDir, ".ccsdb"), 0755)

	db, err := gorm.Open(sqlite.Open(path.Join(u.HomeDir, ".ccsdb", "sqlite.db")), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&model.CcsApplication{})

	db.AutoMigrate(&model.CcsService{})

	db.AutoMigrate(&model.CcsGrid{})

	// Create
	db.Create(&model.CcsService{ServiceId: "SERVICE_1", Name: "SERVICE_1", AppId: "FIRST_APP"})
	db.Create(&model.CcsService{ServiceId: "SERVICE_2", Name: "SERVICE_2", AppId: "FIRST_APP"})
	db.Create(&model.CcsService{ServiceId: "SERVICE_3", Name: "SERVICE_3", AppId: "FIRST_APP"})

	db.Create(&model.CcsApplication{
		AppId:      "FIRST_APP",
		SecretKey:  "FIRST_APP_SECRET_KEY",
		Name:       "FIRST_APP",
		MQEndpoint: "118.195.175.6:9876",
	})

	db.Create(&model.CcsGrid{ParentServiceId: "SERVICE_1", ServiceId: "SERVICE_2"})
	db.Create(&model.CcsGrid{ParentServiceId: "SERVICE_2", ServiceId: "SERVICE_3"})

}

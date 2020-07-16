package main

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/arifseft/clean-architecture-sample/pkg/user"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/spf13/viper"
	"gopkg.in/gormigrate.v1"
)

func main() {
	viper.SetConfigFile("config.json")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s", err.Error())
	}
	mode := viper.GetString("mode")
	dbprefix := "database_" + mode
	dbhost := viper.GetString(dbprefix + ".host")
	dbport := viper.GetString(dbprefix + ".port")
	dbuser := viper.GetString(dbprefix + ".user")
	dbname := viper.GetString(dbprefix + ".dbname")
	dbpassword := viper.GetString(dbprefix + ".password")
	connection := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbuser, dbpassword, dbhost, dbport, dbname)
	val := url.Values{}
	val.Add("parseTime", "true")
	val.Add("loc", "Asia/Jakarta")
	dsn := fmt.Sprintf("%s?%s", connection, val.Encode())
	db, err := gorm.Open(`mysql`, dsn)
	if err != nil {
		log.Fatal(err)
	}
	db.LogMode(true)

	options := gormigrate.Options{
		TableName:                 "go_migrations",
		IDColumnName:              "id",
		IDColumnSize:              255,
		UseTransaction:            false,
		ValidateUnknownMigrations: false,
	}


	m := gormigrate.New(db, &options, []*gormigrate.Migration{
		{
			ID: "2020_07_16_132137_create_table_users",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&user.User{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.DropTable(&user.User{}).Error
			},
		},
	})

	args := os.Args[1:]
	if len(args) > 0 {
		if args[0] == "rollback" {
			if err = m.RollbackLast(); err != nil {
				log.Fatalf("Could not rollback: %v", err)
			}
			log.Printf("RollbackLast did run successfully")
		} else {
			log.Printf("wrong command")
		}
	} else {

		if err = m.Migrate(); err != nil {
			log.Fatalf("Could not migrate: %v", err)
		}
		log.Printf("Migration did run successfully")
	}
}

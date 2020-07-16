package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/arifseft/clean-architecture-sample/api/handler"
	"github.com/arifseft/clean-architecture-sample/api/middleware"
	"github.com/arifseft/clean-architecture-sample/pkg/user"
	"github.com/gorilla/handlers"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)

	viper.SetConfigFile("config.json")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s", err.Error())
	}
	os.Setenv("secret", viper.GetString("jwt_secret"))
}

func dbConnect(host, port, user, dbname, password, sslmode string) (*gorm.DB, error) {

	// In the case of heroku
	// if os.Getenv("DATABASE_URL") != "" {
	//     return gorm.Open("mysql", os.Getenv("DATABASE_URL"))
	// }

	connection := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, dbname)
	val := url.Values{}
	val.Add("parseTime", "true")
	val.Add("loc", "Asia/Jakarta")
	dsn := fmt.Sprintf("%s?%s", connection, val.Encode())
	db, err := gorm.Open(`mysql`, dsn)
	if err != nil {
		log.Fatal(err)
	}

	return db, err
}

func main() {
	mode := viper.GetString("mode")

	// DB binding
	dbprefix := "database_" + mode
	dbhost := viper.GetString(dbprefix + ".host")
	dbport := viper.GetString(dbprefix + ".port")
	dbuser := viper.GetString(dbprefix + ".user")
	dbname := viper.GetString(dbprefix + ".dbname")
	dbpassword := viper.GetString(dbprefix + ".password")
	dbsslmode := viper.GetString(dbprefix + ".sslmode")

	db, err := dbConnect(dbhost, dbport, dbuser, dbname, dbpassword, dbsslmode)
	if err != nil {
		log.Fatalf("Error connecting to the database: %s", err.Error())
	}
	defer db.Close()
	log.Println("Connected to the database")

	// migrations
	// db.AutoMigrate(&user.User{})

	// initializing repos and services
	userRepo := user.NewMysqlRepo(db)

	userSvc := user.NewService(userRepo)

	// Initializing handlers
	r := http.NewServeMux()

	handler.MakeUserHandler(r, userSvc)

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		return
	})

	// HTTP(s) binding
	serverprefix := "server_" + mode
	host := viper.GetString(serverprefix + ".host")
	port := os.Getenv("PORT")
	timeout := time.Duration(viper.GetInt("timeout"))

	if port == "" {
		port = viper.GetString(serverprefix + ".port")
	}

	conn := host + ":" + port

	// middlewares
	mwCors := middleware.CorsEveryWhere(r)
	mwLogs := handlers.LoggingHandler(os.Stdout, mwCors)

	srv := &http.Server{
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		Addr:         conn,
		Handler:      mwLogs,
	}

	log.Printf("Starting in %s mode", mode)
	log.Printf("Server running on %s", conn)
	log.Fatal(srv.ListenAndServe())
}

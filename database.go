package intercom

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/args"
	"github.com/getevo/evo/v2/lib/db"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

var conn *gorm.DB

func setupDatabase() {
	var host = args.Get("-db-host")
	if host == "" {
		host = "127.0.0.1"
	}

	var username = args.Get("-db-username")
	if username == "" {
		username = "root"
	}

	var database = args.Get("-db-database")

	var password = args.Get("-db-password")
	if password == "" {
		password = ""
	}

	var params = args.Get("-db-params")

	var newLog = logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: 200 * time.Millisecond, // Slow SQL threshold
			LogLevel:      logger.Warn,            // Log level
			Colorful:      true,                   // Disable color
		},
	)
	cfg := &gorm.Config{
		Logger: newLog,
	}

	var err error
	connectionString := fmt.Sprintf("%s:%s@tcp(%s)/%s?%s", username, password, host, database, params)
	conn, err = gorm.Open(mysql.Open(connectionString), cfg)
	if err != nil {
		log.Fatal(err)
	}
	db.Register(conn)
}

// GetDBO return database object instance
func GetDBO() *gorm.DB {
	if conn == nil {
		setupDatabase()
	}

	return conn
}

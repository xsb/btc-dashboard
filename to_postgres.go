package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
)

/*
toPostgres() goes through all json files in db-dump,
decodes the file to a their corresponding struct(s), and then inserts them into postgresql tables

If more tables are desired, you would need to add a new db.Insert here,
and define a new model using the new struct definition.
*/
func toPostgres() {
	DB_ADDR, ok := os.LookupEnv("DB_ADDR")
	if !ok {
		DB_ADDR = "http://localhost:5432"
	}

	db := pg.Connect(&pg.Options{
		Addr:     DB_ADDR,
		User:     os.Getenv("DB_USERNAME"),
		Password: os.Getenv("DB_PASSWORD"),
		Database: os.Getenv("DB"),
	})

	defer db.Close()

	model := interface{}((*DashboardDataV2)(nil))
	err := db.CreateTable(model, &orm.CreateTableOptions{
		Temp:        false,
		IfNotExists: true,
	})
	if err != nil {
		fatal("Error creating table: ", err)
	}

	if _, err := os.Stat(JSON_DIR); os.IsNotExist(err) {
		return
	}

	files, err := ioutil.ReadDir(JSON_DIR)
	if err != nil {
		fatal("Error reading directory", err)
	}

	for i, fileInfo := range files {
		file, err := os.Open(JSON_DIR + "/" + fileInfo.Name())
		if err != nil {
			fatal(i, err, fileInfo.Name())
		}

		dec := json.NewDecoder(file)
		var data Data

		err = dec.Decode(&data)
		if err != nil {
			fatal("JSON decoding error: ", err, fileInfo.Name())
		}

		file.Close()

		err = db.Insert(&data.DashboardDataRow)
		if err != nil {
			fatal("Error inserting into db: ", err)
		}

		log.Println("Done with file: ", fileInfo.Name())
	}
}

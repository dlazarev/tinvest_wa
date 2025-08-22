package main

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/png"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"database/sql"

	"github.com/sunshineplan/imgconv"
	_ "modernc.org/sqlite"
)

var db *sql.DB

func initDatabase(dbPath string) error {
	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return nil
	}

	_, err = db.ExecContext(context.Background(),
		`CREATE TABLE IF NOT EXISTS securityLogo (
				figi TEXT PRIMARY KEY,
				url TEXT,
				dateStore DATETIME,
				data BLOB)`,
	)
	if err != nil {
		return err
	}

	return nil
}

//************************************************************************

func logoActual(figi, url string) bool {
	row := db.QueryRowContext(context.Background(),
		`SELECT dateStore FROM securityLogo WHERE figi=?`, figi)

	var dateStore time.Time
	var now = time.Now()

	err := row.Scan(&dateStore)
	if errors.Is(err, sql.ErrNoRows) {
		sql := `INSERT INTO securityLogo (figi, url, dateStore) VALUES (?,?,?);`
		if _, err := db.ExecContext(context.Background(), sql, figi, url, now); err != nil {
			log.Fatalf("Error insert sqlite record: %v", err)
		}
		return false
	}

	futureDate := now.AddDate(0, 0, 28)
	return dateStore.Before(futureDate)
}

//************************************************************************

func makeLogoActual(figi, url string, imageData *image.Image) {
	sql := `UPDATE securityLogo SET figi=?, url=?, dateStore=?, data=?`
	result, err := db.ExecContext(context.Background(), sql, figi, url, time.Now(), imageData)
	if err != nil {
		log.Fatalf("Error update sql table: %v", err)
	}

	fmt.Println(result)
}

//************************************************************************

func updateLogo(acc *AccDetail) {
	for _, sec := range acc.Pos.Securities {
		if sec.InstrumentDesc.Brand.LogoName != "" {
			logoName := strings.Replace(sec.InstrumentDesc.Brand.LogoName, ".png", "x160.png", 1)
			url := logoURL + logoName

			if logoActual(sec.Figi, url) {
				continue
			}

			resp, err := http.Get(url)
			if err != nil {
				log.Fatalf("Error fetching logo: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				log.Fatalf("Bad status code: %d", resp.StatusCode)
			}

			img, _, err := image.Decode(resp.Body)
			if err != nil {
				log.Fatalf("Error decoding image: %v", err)
			}
			mark := imgconv.Resize(img, &imgconv.ResizeOption{Width: 48})
			markFileName := filepath.Join(basePath, imagePath, sec.InstrumentDesc.Brand.LogoName)
			f, err := os.Create(markFileName)
			if err != nil {
				log.Fatalf("Error creating file: %v", err)
			}
			defer f.Close()

			if err = png.Encode(f, mark); err != nil {
				log.Fatalf("Error writing image: %v", err)
			}
			makeLogoActual(sec.Figi, url, &mark)
		}
	}
}

//************************************************************************

package main

import (
	"bytes"
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

	if err = db.Ping(); err != nil {
		return err
	}

	//defer db.Close()

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

func logoActual(figi, url, logoName string) bool {
	row := db.QueryRowContext(context.Background(),
		`SELECT dateStore, data FROM securityLogo WHERE figi=?`, figi)

	var dateStore time.Time
	var now = time.Now()
	var data []byte

	err := row.Scan(&dateStore, &data)
	if errors.Is(err, sql.ErrNoRows) {
		sql := `INSERT INTO securityLogo (figi, url, dateStore) VALUES (?,?,?);`
		if _, err := db.ExecContext(context.Background(), sql, figi, url, now); err != nil {
			log.Fatalf("Error insert sqlite record: %v", err)
		}
		return false
	}

	futureDate := now.AddDate(0, 0, 28)
	if dateStore.Before(futureDate) {
		// Let's check if the logo file is present in the folder.
		logoPath := filepath.Join(basePath, imagePath, logoName)
		if _, err := os.Stat(logoPath); os.IsNotExist(err) {
			if err = os.WriteFile(logoPath, data, 0644); err != nil {
				log.Fatalf("Error writing file logo: %v", err)
			}
		}
	}
	return true
}

//************************************************************************

func makeLogoActual(figi, url string, imageData *image.Image) {
	sql := `UPDATE securityLogo SET url=?, dateStore=?, data=? WHERE figi=?`

	var buf bytes.Buffer
	err := png.Encode(&buf, *imageData)
	if err != nil {
		log.Fatalf("makeLogoActual(). Error encode image data: %v", err)
	}

	result, err := db.ExecContext(context.Background(), sql, url, time.Now(), buf.Bytes(), figi)
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

			if logoActual(sec.Figi, url, sec.InstrumentDesc.Brand.LogoName) {
				continue
			}

			log.Printf("Logo for %v not actual. Download it", sec.InstrumentDesc.Ticker)

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

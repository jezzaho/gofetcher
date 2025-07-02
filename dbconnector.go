package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type Arrival struct {
	RealTime     time.Time
	FlightNumber string
	TTL          int
	Transit      int
}

func connectDB() {
	connStr := ""
	now := time.Now().UTC()
	from := now.AddDate(0, 0, -7)
	to := now

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Cannot open connection: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Cannot connect to DB: %v", err)
	}
	fmt.Println("Connected!")

	query := `
		SELECT 
			cur.real_time AS "ALDT/ATOT",
			btrim(cur.flight_number::text)::character varying AS "Numer rejsu FIDS",
			cur.fis_pax_all AS "TTL",
			cur.fis_pax_transit AS "Transit"
		FROM 
			arrivals_curr cur
		LEFT JOIN 
			arrivals_departures_curr adc 
			ON adc.arrival_curr_id = cur.id 
			AND adc.delete_status::smallint >= 0
		WHERE 
			cur.delete_status::smallint >= 0
			AND adc.id IS NOT NULL
			AND cur.real_time < $1
			AND cur.real_time >= $2
		ORDER BY 
			"ALDT/ATOT" ASC
		`

	rows, err := db.Query(query, to, from)
	if err != nil {
		log.Fatalf("Query error: %v", err)
	}
	defer rows.Close()

	var Arrresults []Arrival

	for rows.Next() {
		var a Arrival
		if err := rows.Scan(&a.RealTime, &a.FlightNumber, &a.TTL, &a.Transit); err != nil {
			log.Fatalf("Row scan error: %v", err)
		}
		Arrresults = append(Arrresults, a)
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("Rows error: %v", err)
	}

}

func 

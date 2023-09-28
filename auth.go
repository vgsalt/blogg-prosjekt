package main

import (
	"log/slog"
	"net/http"
)

func LoginPage(w http.ResponseWriter, r *http.Request) {
	// Hvis brukeren bruker GET for å hente nettsiden, så viser vi nettsiden for å logge in.
	if r.Method == "GET" {
		if err := Tmpl.ExecuteTemplate(w, "login.html", nil); err != nil {
			slog.Error("Kunne ikke vise side for å logge inn", "feil", err)
			w.Write([]byte("Kunne ikke vise side for å logge inn. Sjekk konsoll for mer informasjon."))
		}
	}
	// Ellers med POST så bruker vi koden for å logge inn fra bakenden. GET er for å hente, POST er for å sende i HTTP.
	if r.Method == "POST" {
		w.Write([]byte("etterpå"))
	}
}

// blogg-prosjekt er et prosjekt som lar deg ... noe
package main

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/microcosm-cc/bluemonday"
)

var funcMap = template.FuncMap{
	"ConvertToUtc": func(epoch int) time.Time {
		utcTime := time.Unix(int64(epoch), 0)
		return utcTime
	},
	"sanitize": Sanitize,
	"toHtml": func(html string) template.HTML {
		return template.HTML(html)
	},
}

var Tmpl = template.Must(template.New("artikler").Funcs(funcMap).ParseFS(TemplateFolder, "templates/*"))

func main() {
	// Si ifra til terminalen at bloggen starter.
	slog.Info("Starter blogg.")

	// Last inn miljøvariabler fra en .env fil.
	err := godotenv.Load()
	if err != nil {
		slog.Error("Noe gikk galt med å laste inn .env filen.")
	}

	// Koble til PostgreSQL, med URL-en inne i miljøvariablene til PC-en (eller .env)
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.Error("Noe gikk galt med å koble til Postgres :(", "error", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	r := chi.NewRouter()

	// Index-siden.
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// Spør SQL om alle artiklene.
		rows, err := conn.Query(context.Background(), `SELECT * FROM artikler ORDER BY dato DESC`)
		if err != nil {
			slog.Error("Noe gikk galt mens vi prøvde å hente artikler fra database.", "feil", err)
			// Si ifra til brukeren at noe gikk galt.
			w.Write([]byte("Funker ikke :("))
			return
		}
		defer rows.Close()

		var artikler []Article

		// For hver rad fra sql-queryen, gjør dette
		for rows.Next() {
			var a Article
			// Les inn verdiene fra radene inn i typen våres
			err := rows.Scan(&a.Tittel, &a.Dato, &a.Forfatter, &a.Innhold)
			if err != nil {
				slog.Error("Noe gikk galt med å skanne artiklene", "feil", err)
			}
			// Skriv inn typene inn i arrayen våres
			artikler = append(artikler, a)
		}

		if err := rows.Err(); err != nil {
			slog.Error("Kunne ikke hente artikler fra database.", "feil", err)
			panic(err)
		}

		// Informasjon som skal vises til brukeren
		p := Page{
			Title:    "Testblogg",
			Artikler: artikler,
		}
		if err := Tmpl.ExecuteTemplate(w, "index.html", p); err != nil {
			slog.Error("Oops! Her gikk det noe galt.", err)
			return
		}
	})

	r.Get("/post", func(w http.ResponseWriter, r *http.Request) {
		if err := Tmpl.ExecuteTemplate(w, "post.html", nil); err != nil {
			slog.Error("Oops! Her gikk det noe galt.", err)
			return
		}
	})

	r.Post("/post", func(w http.ResponseWriter, r *http.Request) {
		// få tak i tiden akkurat nå i epoch
		currentTime := time.Now().Unix()
		// si til databasen at vi vil lage en ny artikkel
		_, err := conn.Query(context.Background(), `INSERT INTO artikler (tittel, dato, forfatter, innhold) VALUES ($1, $2, $3, $4)`, r.FormValue("tittel"), int(currentTime), r.FormValue("forfatter"), Sanitize(template.HTML(r.FormValue("innhold"))))
		if err != nil {
			slog.Error("Kunne ikke skrive inn artikkel inn i database", "error", err)
		}

		// send brukeren til artikkelen
		http.Redirect(w, r, "/artikler/"+r.FormValue("tittel"), http.StatusSeeOther)
	})

	r.Get("/artikkel/{artikkel}", func(w http.ResponseWriter, r *http.Request) {
		// lagre artikkelen brukeren spør om som variabel
		artikkel := chi.URLParam(r, "artikkel")

		var ap ArticlePage

		// Spør om artikler som matcher artikkelen som brukeren spør om
		err := conn.QueryRow(context.Background(), "select * from artikler where tittel=$1", artikkel).Scan(&ap.Tittel, &ap.Dato, &ap.Forfatter, &ap.Innhold)
		if err != nil {
			slog.Error("Kunne ikke finne artikkelen!", "feil", err)
			w.Write([]byte("Kunne ikke finne artikkelen. Mer info tilgjengelig i loggen."))
			return
		}

		// sett Title til Tittel (rart, ikke sant?)
		ap.Title = ap.Tittel

		if err := Tmpl.ExecuteTemplate(w, "article.html", ap); err != nil {
			slog.Error("Oops! Her gikk det noe galt.", "error", err)
			panic(err)
		}
	})

	// For filer som css, font osv.
	r.Handle("/static/*", http.FileServer(http.FS(StaticFolder)))

	r.Get("/login", LoginPage)
	r.Post("/login", LoginPage)

	// forklar hvor serveren skal kjøres. kunne ha brukt egen variabel for dette men
	// http.Server ser litt finere ut kanskje?

	server := http.Server{
		Addr: os.Getenv("SERVE_AT"),
	}

	slog.Info("Blogg er oppe!", "adresse", server.Addr)

	http.ListenAndServe(server.Addr, r)
}

func Sanitize(html template.HTML) string {
	p := bluemonday.UGCPolicy()

	sanitized := p.Sanitize(string(html))

	return sanitized
}

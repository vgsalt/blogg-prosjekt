package main

import (
	"embed"
	"html/template"
)

var (
	//go:embed templates/*
	TemplateFolder embed.FS
	//go:embed static/*
	StaticFolder embed.FS
)

type Page struct {
	Title    string
	Artikler []Article
}

type ArticlePage struct {
	Title     string
	Tittel    string
	Dato      int
	Forfatter string
	Innhold   template.HTML
}

type Article struct {
	Tittel    string
	Dato      int
	Forfatter string
	Innhold   string
}

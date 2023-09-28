package main

import (
	"embed"
)

var (
	//go:embed templates/*
	TemplateFolder embed.FS
	//go:embed static/*
	StaticFolder embed.FS
)

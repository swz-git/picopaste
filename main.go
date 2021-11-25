package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"git.mills.io/prologic/bitcask"
	"github.com/fatih/color"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/lucasjones/reggen"
)

//go:embed html/*
var content embed.FS

//go:embed asciilogo.txt
var asciilogo string

type PasteInfo struct {
	Name  string
	Value string
}

func GenID(db *bitcask.Bitcask) string {
	regex := "^[A-Za-z0-9_\\-\\.~]{8}$"
	id, _ := reggen.Generate(regex, 8)
	for db.Has([]byte(id)) {
		id, _ = reggen.Generate(regex, 8)
	}
	return id
}

func main() {
	data, _ := content.ReadFile("html/index.html")
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	// Setup badger db
	log.Println("Opening database...")
	dbpath := os.Getenv("PICOPASTE_DB_PATH")
	if dbpath == "" {
		dbpath = "/tmp/db"
	}
	db, err := bitcask.Open(dbpath)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Database opened.")

	app.Use(limiter.New(limiter.Config{
		Next: func(c *fiber.Ctx) bool {
			return c.Method() == "GET"
		},
		Max:        2,
		Expiration: 30 * time.Second,
	}))

	// Logger
	app.Get("/*", func(c *fiber.Ctx) error {
		log.Println(c.Path())
		c.Next()
		return nil
	})

	// Handle raw requests
	app.Get("/raw/:id", func(c *fiber.Ctx) error {
		c.Context().SetContentType("text/plain")
		value, err := db.Get([]byte(c.Params("id")))
		if err != nil {
			return c.SendStatus(http.StatusNotFound)
		}
		content := string(value)
		return c.SendString(content)
	})

	// Handle normal requests
	app.Get("/:id", func(c *fiber.Ctx) error {
		c.Context().SetContentType("text/html")

		value, err := db.Get([]byte(c.Params("id")))
		if err != nil {
			if db.Has([]byte(c.Params("id"))) {
				log.Fatal(err)
			} else {
				return c.Redirect("/")
			}
		}

		tmpl, err := template.New("Page").Parse(string(data))
		if err != nil {
			panic(err)
		}

		u := PasteInfo{
			Name:  c.Params("id"),
			Value: string(value),
		}

		var temp bytes.Buffer

		erro := tmpl.Execute(&temp, u)
		if erro != nil {
			log.Fatal(erro)
		}

		return c.SendString(temp.String())
	})

	// Default
	app.Get("/", func(c *fiber.Ctx) error {
		c.Context().SetContentType("text/html")

		tmpl, err := template.New("Page").Parse(string(data))
		if err != nil {
			panic(err)
		}

		u := PasteInfo{
			Name:  "Quick pasting service",
			Value: "",
		}

		var temp bytes.Buffer

		erro := tmpl.Execute(&temp, u)
		if erro != nil {
			log.Fatal(erro)
		}

		return c.SendString(temp.String())
	})

	app.Post("/paste", func(c *fiber.Ctx) error {
		c.Context().SetContentType("text/plain")

		id := GenID(db)
		data := c.Body()
		if len(data) == 0 {
			return c.SendStatus(http.StatusBadRequest)
		}

		exists := ""
		db.Fold(func(key []byte) error {
			value, _ := db.Get(key)
			if string(value) == string(data) {
				exists = string(key)
				return nil
			} else {
				return nil
			}
		})
		if exists != "" {
			return c.SendString(fmt.Sprintf("/%s", exists))
		}

		err := db.Put([]byte(id), []byte(string(data)))
		if err != nil {
			log.Fatal(err)
		}

		return c.SendString(fmt.Sprintf("/%s", id))
	})

	app.Use("/public", filesystem.New(filesystem.Config{
		Root:       http.FS(content),
		PathPrefix: "html",
	}))

	port := os.Getenv("PICOPASTE_PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println(asciilogo)
	color.Green("Picopaste is running on port " + port + " (http://127.0.0.1:" + port + ")")
	app.Listen(":" + port)
}

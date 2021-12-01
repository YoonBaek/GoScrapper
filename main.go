package main

import (
	"os"
	"strings"

	"github.com/YoonBaek/GoScrapper/scrapper"
	"github.com/labstack/echo"
)

const FILENAME = "jobs.csv"

func handleHome(c echo.Context) error {
	return c.File("home.html")
}

func handleScrap(c echo.Context) error {
	defer os.Remove(FILENAME)
	searchWord := strings.ToLower(scrapper.Stringstidy(c.FormValue("searchWord")))
	scrapper.Scrap(searchWord)
	return c.Attachment(FILENAME, FILENAME)
}

func main() {
	e := echo.New()
	e.GET("/", handleHome)
	e.POST("/scrap", handleScrap)
	e.Logger.Fatal(e.Start(":1323"))
	// scrapper.Scrap()
}

package rest

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gyujae/jobscrapper_backend/scrapper"
)

type SiteImage struct {
	Site string `json:"site"`
	Url  string `json:"url"`
}

func Start(port string) {
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowHeaders: "Access-Control-Allow-Origin",
	}))

	app.Get("/sites/images", handleSitesImage)

	app.Get("/sites", handleSites)

	app.Get("/:keyword", handleKeyword)

	log.Fatal(app.Listen(port))
}

func handleSitesImage(c *fiber.Ctx) error {
	var result []SiteImage
	for site, url := range scrapper.WebsitesImages {
		item := SiteImage{Site: site, Url: url}
		result = append(result, item)
	}
	if len(scrapper.WebsitesImages) == 0 {
		return c.Status(404).JSON(&fiber.Map{
			"ok":    false,
			"sites": nil,
			"error": "There are no Site!",
		})
	}

	return c.JSON(&fiber.Map{
		"ok":    true,
		"sites": result,
		"error": nil,
	})
}

func handleSites(c *fiber.Ctx) error {
	var sites []string
	for key := range scrapper.Websites {
		sites = append(sites, key)
	}
	if len(scrapper.Websites) == 0 {
		return c.Status(404).JSON(&fiber.Map{
			"ok":    false,
			"sites": nil,
			"error": "There are no Site!",
		})
	}
	return c.JSON(&fiber.Map{
		"ok":    true,
		"sites": sites,
		"error": nil,
	})
}

func handleKeyword(c *fiber.Ctx) error {
	keyword := c.Params("keyword")
	jobs := scrapper.SplitJobsBySite(keyword)
	if len(jobs) == 0 {
		return c.Status(404).JSON(&fiber.Map{
			"ok":        false,
			"error":     "There are no jobs!",
			"jobs":      nil,
			"totalJobs": 0,
		})
	}

	totalJobs := 0
	for _, siteJobs := range jobs {
		totalJobs += len(siteJobs)
	}
	return c.JSON(&fiber.Map{
		"ok":        true,
		"jobs":      jobs,
		"totalJobs": totalJobs,
		"error":     nil,
	})
}

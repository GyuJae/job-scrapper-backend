package rest

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gyujae/jobscrapper_backend/scrapper"
)

func Start(port string) {
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowHeaders: "Access-Control-Allow-Origin",
	}))

	app.Get("/sites", func(c *fiber.Ctx) error {
		var sites []string
		for key := range scrapper.Websites {
			sites = append(sites, key)
		}
		if len(scrapper.Websites) == 0 {
			return c.Status(404).JSON(&fiber.Map{
				"ok":    false,
				"sites": nil,
				"error": "There are no jobs!",
			})
		}
		return c.JSON(&fiber.Map{
			"ok":    true,
			"sites": sites,
			"error": nil,
		})
	})

	app.Get("/:keyword", func(c *fiber.Ctx) error {
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
	})

	log.Fatal(app.Listen(port))
}

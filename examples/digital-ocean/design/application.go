package design

import . "github.com/raphael/goa/design"

var _ = Application("gdo", func() {

	Resource("droplets", func() {
		Action("index", func() {
			Routing(Get(""))
			Description("List all droplets")
		})

		Action("create", func() {
			Routing(Post(""))
			Description("Create a new droplet")
			Payload(
				Attribute("name", String, func() {
					Description("The human-readable string you wish to use when displaying the Droplet name. The name, if set to a domain name managed in the DigitalOcean DNS management system, will configure a PTR record for the Droplet. The name set during creation will also determine the hostname for the Droplet in its internal configuration.")
					Required()
				}),
				Attribute("region", String, func() {
					Description("The unique slug identifier for the region that you wish to deploy in.")
					Required()
				}),
				Attribute("size", String, func() {
					Description("The unique slug identifier for the size that you wish to select for this Droplet.")
					Required()
				}),
				Attribute("imageID", Integer, func() {
					Description("The image ID of a public or private image. One and exactly one of ImageID or ImageSlug must be specified. This image will be the base image for your Droplet.")
				}),
				Attribute("imageSlug", String, func() {
					Description("The unique slug identifier for a public image. One and exactly one of ImageID or ImageSlug must be specified. This image will be the base image for your Droplet.")
				}),
				Attribute("sshKeys", CollectionOf(Integer), func() {
					Description("An array containing the IDs of the SSH keys that you wish to embed in the Droplet's root account upon creation.")
				}),
				Attribute("backups", Bool, func() {
					Description("A boolean indicating whether automated backups should be enabled for the Droplet. Automated backups can only be enabled when the Droplet is created.")
				}),
				Attribute("ipv6", Bool, func() {
					Description("A boolean indicating whether IPv6 is enabled on the Droplet.")
				}),
				Attribute("privateNetworking", Bool, func() {
					Description("A boolean indicating whether private networking is enabled for the Droplet. Private networking is currently only available in certain regions.")
				}),
				Attribute("userData", String, func() {
					Description("A string of the desired User Data for the Droplet. User Data is currently only available in regions with metadata listed in their features.")
				}),
			)
			Response(NoContent)
		})
	})

	Resource("images", func() {
		Action("index", func() {
			Routing(Get(""))
			Description("List all images optionally filtered by type or private")
			Params(
				Attribute("type", String, func() {
					Enum("distribution", "application")
				}),
				Attribute("private", Bool),
			)
			Response(Ok, func() {
				MediaType(CollectionOf(ImageMediaType))
			})
		})

	})
})

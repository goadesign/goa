package design

import . "github.com/raphael/goa/design"

// Droplet media type
var DropletMediaType = MediaType("application/vnd.goa.examples.do.droplet", func() {
	Description("A Droplet is a DigitalOcean virtual machine.")
	Attributes(
		Attribute("id", Integer, func() {
			Description("A unique identifier for each Droplet instance. This is automatically generated upon Droplet creation.")
		}),
		Attribute("name", String, func() {
			Description("The human-readable name set for the Droplet instance.")
		}),
		Attribute("memory", Integer, func() {
			Description("Memory of the Droplet in megabytes.")
		}),
		Attribute("vcpus", Integer, func() {
			Description("The number of virtual CPUs.")
		}),
		Attribute("disk", Integer, func() {
			Description("The size of the Droplet's disk in gigabytes.")
		}),
		Attribute("locked", Integer, func() {
			Description("A boolean value indicating whether the Droplet has been locked, preventing actions by users.")
		}),
		Attribute("created_at", String, func() {
			Description("A time value given in ISO8601 combined date and time format that represents when the Droplet was created.")
		}),
		Attribute("status", String, func() {
			Description("A status string indicating the state of the Droplet instance. This may be \"new\", \"active\", \"off\", or \"archive\".")
		}),
		Attribute("backup_ids", CollectionOf(Integer), func() {
			Description("An array of backup IDs of any backups that have been taken of the Droplet instance. Droplet backups are enabled at the time of the instance creation.")
		}),
		Attribute("snapshot_ids", CollectionOf(Integer), func() {
			Description("An array of snapshot IDs of any snapshots created from the Droplet instance.")
		}),
		Attribute("features", CollectionOf(String), func() {
			Description("An array of features enabled on this Droplet.")
		}),
		Attribute("region", Region, func() {
			Description("The region that the Droplet instance is deployed in. When setting a region, the value should be the slug identifier for the region. When you query a Droplet, the entire region object will be returned.")
		}),
		Attribute("image", Image, func() {
			Description("The base image used to create the Droplet instance. When setting an image, the value is set to the image id or slug. When querying the Droplet, the entire image object will be returned.")
		}),
		Attribute("size", Size, func() {
			Description("The current size object describing the Droplet. When setting a size, the value is set to the size slug. When querying the Droplet, the entire size object will be returned. Note that the disk volume of a droplet may not match the size's disk due to Droplet resize actions. The disk attribute on the Droplet should always be referenced.")
		}),
		Attribute("size_slug", String, func() {
			Description("The unique slug identifier for the size of this Droplet.")
		}),
		Attribute("networks", Networks, func() {
			Description("The details of the network that are configured for the Droplet instance. This is an object that contains keys for IPv4 and IPv6. The value of each of these is an array that contains objects describing an individual IP resource allocated to the Droplet. These will define attributes like the IP address, netmask, and gateway of the specific network depending on the type of network it is.")
		}),
		Attribute("kernel", Kernel, func() {
			Description("The current kernel. This will initially be set to the kernel of the base image when the Droplet is created.")
		}),
		Attribute("next_backup_window", BackupWindow, func() {
			Description("The details of the Droplet's backups feature, if backups are configured for the Droplet. This object contains keys for the start and end times of the window during which the backup will start.")
		}),
	)
})

var Region = Type("region", func() {
	Attributes(
		Attribute("name", String, func() {
			Description("The display name of the region. This will be a full name that is used in the control panel and other interfaces.")
		}),
		Attribute("slug", String, func() {
			Description("A human-readable string that is used as a unique identifier for each region.")
		}),
		Attribute("sizes", CollectionOf(String), func() {
			Description("This attribute is set to an array which contains the identifying slugs for the sizes available in this region.")
		}),
		Attribute("available", Bool, func() {
			Description("This is a boolean value that represents whether new Droplets can be created in this region.")
		}),
		Attribute("features", CollectionOf(String), func() {
			Description("This attribute is set to an array which contains features available in this region")
		}),
	)
})

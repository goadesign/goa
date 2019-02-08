package design

import . "goa.design/goa/dsl"

var _ = Service("resume", func() {
	Description("The storage service makes it possible to add resumes using multipart.")

	HTTP(func() {
		Path("/resumes")
	})

	Method("list", func() {
		Description("List all stored resumes")
		Result(CollectionOf(StoredResume))
		HTTP(func() {
			GET("/")
			Response(StatusOK)
		})
	})

	Method("add", func() {
		Description("Add n number of resumes and return their IDs. This is a multipart request and each part has field name 'resume' and contains the encoded resume to be added.")
		Payload(ArrayOf(Resume))
		Result(ArrayOf(Int))
		HTTP(func() {
			POST("/")
			MultipartRequest()
		})
	})
})

var Resume = Type("Resume", func() {
	Attribute("name", String, "Name in the resume")
	Attribute("experience", ArrayOf(Experience), "Experience section in the resume")
	Attribute("education", ArrayOf(Education), "Education section in the resume")
	Required("name")
})

var Experience = Type("Experience", func() {
	Attribute("company", String, "Name of the company")
	Attribute("role", String, "Name of the role in the company")
	Attribute("duration", Int, "Duration (in years) in the company")
	Required("company", "role", "duration")
})

var Education = Type("Education", func() {
	Attribute("institution", String, "Name of the institution")
	Attribute("major", String, "Major name")
	Required("institution", "major")
})

var StoredResume = ResultType("application/vnd.goa.resume", func() {
	TypeName("StoredResume")
	Attributes(func() {
		Extend(Resume)
		Attribute("id", Int, "ID of the resume")
		Attribute("created_at", String, "Time when resume was created")
		Required("id", "name", "experience", "education", "created_at")
	})
	View("default", func() {
		Attribute("id")
		Attribute("name")
		Attribute("experience")
		Attribute("education")
		Attribute("created_at")
	})
})

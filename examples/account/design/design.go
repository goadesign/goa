package design

import . "goa.design/goa.v2/design/rest"
import . "goa.design/goa.v2/dsl/rest"

var _ = API("basic", func() {
	Title("Basic REST example")
	Description("A simple example to  help build v2")
	Server("http://localhost:8080")
})

var _ = Service("account", func() {
	Description("Manage accounts")
	HTTP(func() {
		Path("/orgs/{org_id}/accounts")
		Param("org_id", UInt, "ID of owner organization", func() {
			Maximum(10000)
			Example("basic", 123)
		})
	})
	Endpoint("create", func() {
		Description("Create new account")
		Payload(CreateAccount)
		Result(Account)
		Error("name_already_taken", NameAlreadyTaken, "Error returned when name is already taken")
		HTTP(func() {
			POST("/")
			Response(StatusCreated, func() {
				Header("Href:Location")
			})
			Response(StatusAccepted, func() {
				Header("Href:Location")
				Body(Empty)
			})
			Response("name_already_taken", StatusConflict)
		})
	})
	Endpoint("list", func() {
		Description("List all accounts")
		Payload(ListAccount)
		Result(ArrayOf(Account))
		HTTP(func() {
			GET("/")
			Param("filter")
		})
	})
	Endpoint("show", func() {
		Description("Show account by ID")
		Payload(func() {
			Attribute("org_id", UInt, "ID of organization that owns  account")
			Attribute("id", String, "ID of account to show")
			Example("basic", Val{"org_id": 123, "id": "account1"})
		})
		Result(Account)
		HTTP(func() {
			GET("/{id}")
		})
	})
	Endpoint("delete", func() {
		Description("Delete account by IF")
		Payload(func() {
			Attribute("org_id", UInt, "ID of organization that owns  account")
			Attribute("id", String, "ID of account to show")
			Example("basic", Val{"org_id": 123, "id": "account1"})
		})
		Result(Empty)
		HTTP(func() {
			DELETE("/{id}")
		})
	})
})

var CreateAccount = Type("CreateAccount", func() {
	Description("CreateAccount is the account creation payload")
	Attribute("org_id", UInt, "ID of organization that owns newly created account")
	Attribute("name", String, "Name of new account")
	Attribute("description", String, "Description of new account")
	Required("org_id", "name")
})

var ListAccount = Type("ListAccount", func() {
	Description("ListAccount is the list account payload, it defines an optional list filter")
	Attribute("org_id", UInt, "ID of organization that owns newly created account")
	Attribute("filter", String, "Filter is the account name prefix filter", func() {
		Example("prefix", "go")
	})
})

var Account = MediaType("application/vnd.basic.account", func() {
	TypeName("Account")
	Description("Account type")
	Reference(CreateAccount)
	Attributes(func() {
		Attribute("href", String, "Href to account")
		Attribute("id", String, "ID of account")
		Attribute("org_id")
		Attribute("name")
		Attribute("description", func() {
			Default("An active account")
		})
		Required("href", "id", "org_id", "name")
	})
})

var NameAlreadyTaken = Type("NameAlreadyTaken", func() {
	Description("NameAlreadyTaken is the type returned when creating an account fails because its name is already taken")
	Attribute("message", String, "Message of error")
	Required("message")
})

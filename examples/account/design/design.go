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
			Example(123)
		})
	})
	Method("create", func() {
		Description("Create new account")
		Payload(CreatePayload)
		Result(Account)
		Error("name_already_taken", NameAlreadyTaken, "Error returned when name is already taken")
		HTTP(func() {
			POST("/")
			Response(StatusCreated, func() {
				Header("href:Location")
			})
			Response(StatusAccepted, func() {
				Header("href:Location")
				Body(Empty)
				Tag("status", "provisioning")
			})
			Response("name_already_taken", StatusConflict)
		})
	})
	Method("list", func() {
		Description("List all accounts")
		Payload(ListPayload)
		Result(ArrayOf(Account))
		HTTP(func() {
			GET("/")
			Param("filter")
		})
	})
	Method("show", func() {
		Description("Show account by ID")
		Payload(func() {
			Attribute("org_id", UInt, "ID of organization that owns  account")
			Attribute("id", String, "ID of account to show")
			Required("org_id", "id")
			Example(Val{"org_id": 123, "id": "account1"})
		})
		Result(Account)
		Error("not_found", NotFound, "Account not found")
		HTTP(func() {
			GET("/{id}")
			Response("not_found", StatusNotFound)
		})
	})
	Method("delete", func() {
		Description("Delete account by IF")
		Payload(func() {
			Attribute("org_id", UInt, "ID of organization that owns  account")
			Attribute("id", String, "ID of account to show")
			Required("org_id", "id")
			Example(Val{"org_id": 123, "id": "account1"})
		})
		Result(Empty)
		Error("not_found", NotFound, "Account not found")
		HTTP(func() {
			DELETE("/{id}")
		})
	})
})

var CreatePayload = Type("CreatePayload", func() {
	Description("CreatePayload is the account creation payload")
	Attribute("org_id", UInt, "ID of organization that owns newly created account")
	Attribute("name", String, "Name of new account")
	Attribute("description", String, "Description of new account")
	Required("org_id", "name")
})

var ListPayload = Type("ListPayload", func() {
	Description("ListPayload is the list account payload, it defines an optional list filter")
	Attribute("org_id", UInt, "ID of organization that owns newly created account")
	Attribute("filter", String, "Filter is the account name prefix filter", func() {
		Example("prefix", "go")
	})
})

var Account = ResultType("application/vnd.basic.account", func() {
	TypeName("Account")
	Description("Account type")
	Reference(CreatePayload)
	Attributes(func() {
		Attribute("href", String, "Href to account")
		Attribute("id", String, "ID of account")
		Attribute("org_id")
		Attribute("name")
		Attribute("description", func() {
			Default("An active account")
		})
		Attribute("status", String, "Status of account", func() {
			Enum("provisioning", "ready", "deprovisioning")
		})
		Required("href", "id", "org_id", "name")
	})
})

var NameAlreadyTaken = Type("NameAlreadyTaken", func() {
	Description("NameAlreadyTaken is the type returned when creating an account fails because its name is already taken")
	Attribute("message", String, "Message of error")
	Required("message")
})

var NotFound = Type("NotFound", func() {
	Description("NotFound is the type returned when attempting to show or delete an account that does not exist.")
	Attribute("message", String, "Message of error", func() {
		Example("account 1 of organization 2 not found")
	})
	Attribute("org_id", UInt, "ID of missing account owner organization", func() {
		Example(123)
	})
	Attribute("id", String, "ID of missing account", func() {
		Example("1")
	})
	Required("message", "org_id", "id")
})

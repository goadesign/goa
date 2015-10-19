/*
Package genschema provides a generator for the JSON schema controller.
The schema controller responds to GET /schema requests with the API JSON Hyper-schema.
This JSON schema can be used to generate API documentation, ruby and Go API clients.
See the blog post (https://blog.heroku.com/archives/2014/1/8/json_schema_for_heroku_platform_api)
describing how Heroku leverages the JSON Hyper-schema standard (http://json-schema.org/latest/json-schema-hypermedia.html)
for more information.
*/
package genschema

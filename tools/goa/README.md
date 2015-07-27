goa
===

`goa` is the goa code generator tool. It analyze a goa application metadata and generates the code
for the following components:
* Resources: struct types describing the API resources and action payloads.
* Handlers: functions that decode the request parameters, call the user code and encode the 
            responses.
* Contexts: wrappers around the request parameters that provide typed access and helper functions
            to generate the response contents.

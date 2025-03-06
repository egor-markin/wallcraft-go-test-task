package config

const (
	ApiPrefix          = "/api/v1"
	ProductsApiPrefix  = ApiPrefix + "/products"
	CustomersApiPrefix = ApiPrefix + "/customers"
	InvoicesApiPrefix  = ApiPrefix + "/invoices"

	ContentTypeJSON        = "application/json"
	InternalServerErrorMsg = "Internal server error"
	MethodNotAllowedMsg    = "Method not allowed"

	DefaultServiceBindingAddress = "0.0.0.0:8080"
)

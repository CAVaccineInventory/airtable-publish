package v1

type Metadata struct {
	ApiVersion  ApiVersion
	Contact     Contact
	UsageNotice string
}

type ApiVersion struct {
	Major int
	Minor int
	Stability string
}

type Contact struct {
	PartnersEmail string
}

var DefaultContact = Contact{
	PartnersEmail: "partners@vaccinateca.com",
}
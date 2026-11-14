package domain

// ErrExternalProductNotFound means the external catalog (Open Food Facts)
// has nothing for a barcode. It never reaches the API: a miss there just
// means no suggestion to prefill a new item, not a failure.
var ErrExternalProductNotFound = &Error{Kind: KindNotFound, Code: "external_product_not_found", Message: "El producto no está en el catálogo externo"}

// ExternalProduct is what a public catalog knows about a barcode, offered
// only as a suggestion to prefill a new item: the person creating it can
// change anything before saving.
type ExternalProduct struct {
	Name string
	// Description is nil when the catalog has neither a brand nor a
	// package size for the product.
	Description *string
}

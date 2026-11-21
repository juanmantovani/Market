package postgres

import (
	"context"
	"errors"
	"slices"
	"strings"
	"unicode"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
	repoid "market/backend/internal/infrastructure/repository"
)

var _ gateway.CustomerRepository = (*CustomerRepository)(nil)

const customerDocumentKeyConstraint = "customers_document_key"

type CustomerRepository struct {
	pool *pgxpool.Pool
}

func NewCustomerRepository(pool *pgxpool.Pool) *CustomerRepository {
	return &CustomerRepository{pool: pool}
}

func (repository *CustomerRepository) Create(ctx context.Context, customer domain.Customer) (domain.Customer, error) {
	customer.ID = repoid.NewID()

	_, err := querierFrom(ctx, repository.pool).Exec(ctx, `
		INSERT INTO customers (id, name, document_type, document_number, tax_condition, email, phone, address, notes, is_system, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, false, $10, $10)`,
		customer.ID, customer.Name, customer.DocumentType, customer.DocumentNumber, customer.TaxCondition,
		customer.Email, customer.Phone, customer.Address, customer.Notes, customer.CreatedAt)
	if isUniqueViolation(err, customerDocumentKeyConstraint) {
		return domain.Customer{}, domain.ErrCustomerDocumentInUse
	}
	if err != nil {
		return domain.Customer{}, err
	}

	customer.UpdatedAt = customer.CreatedAt
	return customer, nil
}

func (repository *CustomerRepository) Get(ctx context.Context, id string) (domain.Customer, error) {
	row := querierFrom(ctx, repository.pool).QueryRow(ctx, customerSelect+" WHERE id = $1", id)

	customer, err := scanCustomer(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Customer{}, domain.ErrCustomerNotFound
	}
	return customer, err
}

func (repository *CustomerRepository) Update(ctx context.Context, customer domain.Customer) (domain.Customer, error) {
	tag, err := querierFrom(ctx, repository.pool).Exec(ctx, `
		UPDATE customers
		SET name = $2, document_type = $3, document_number = $4, tax_condition = $5,
		    email = $6, phone = $7, address = $8, notes = $9, updated_at = $10
		WHERE id = $1`,
		customer.ID, customer.Name, customer.DocumentType, customer.DocumentNumber, customer.TaxCondition,
		customer.Email, customer.Phone, customer.Address, customer.Notes, customer.UpdatedAt)
	if isUniqueViolation(err, customerDocumentKeyConstraint) {
		return domain.Customer{}, domain.ErrCustomerDocumentInUse
	}
	if err != nil {
		return domain.Customer{}, err
	}
	if tag.RowsAffected() == 0 {
		return domain.Customer{}, domain.ErrCustomerNotFound
	}

	return customer, nil
}

// Delete removes a customer for good. Once the Sale adapter arrives
// (repository/postgres sale_repository.go), it must also fail with
// domain.ErrCustomerInUse while a sale names the customer, as the memory
// adapter and the SaleRepository contract already require.
func (repository *CustomerRepository) Delete(ctx context.Context, id string) error {
	tag, err := querierFrom(ctx, repository.pool).Exec(ctx, "DELETE FROM customers WHERE id = $1", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrCustomerNotFound
	}

	return nil
}

// List fetches every customer and filters, sorts and pages it in Go so its
// behavior matches the memory adapter's accent- and separator-insensitive
// search exactly. A customers table is small enough for a shop POS that this
// costs nothing worth optimizing with SQL-side search today.
func (repository *CustomerRepository) List(ctx context.Context, filter gateway.CustomerFilter) (domain.Page[domain.Customer], error) {
	page := domain.Page[domain.Customer]{Items: []domain.Customer{}, Page: filter.Page.Page, Limit: filter.Page.Limit}

	rows, err := querierFrom(ctx, repository.pool).Query(ctx, customerSelect)
	if err != nil {
		return domain.Page[domain.Customer]{}, err
	}
	defer rows.Close()

	var all []domain.Customer
	for rows.Next() {
		customer, err := scanCustomer(rows)
		if err != nil {
			return domain.Page[domain.Customer]{}, err
		}
		all = append(all, customer)
	}
	if err := rows.Err(); err != nil {
		return domain.Page[domain.Customer]{}, err
	}

	search := domain.NormalizeText(filter.Search)
	searchDigits := strings.Map(keepDigit, filter.Search)
	var matches []domain.Customer
	for _, customer := range all {
		if customerMatches(customer, search, searchDigits) {
			matches = append(matches, customer)
		}
	}
	slices.SortFunc(matches, func(a, b domain.Customer) int {
		if normalized := strings.Compare(domain.NormalizeText(a.Name), domain.NormalizeText(b.Name)); normalized != 0 {
			return normalized
		}
		return strings.Compare(a.ID, b.ID)
	})

	page.Total = len(matches)
	start := min(filter.Page.Offset(), len(matches))
	end := min(start+filter.Page.Limit, len(matches))
	page.Items = append(page.Items, matches[start:end]...)

	return page, nil
}

const customerSelect = `
	SELECT id, name, document_type, document_number, tax_condition, email, phone, address, notes, is_system, created_at, updated_at
	FROM customers`

// row is what both pgx.Row and pgx.Rows offer to scanCustomer.
type row interface {
	Scan(dest ...any) error
}

func scanCustomer(row row) (domain.Customer, error) {
	var customer domain.Customer
	err := row.Scan(
		&customer.ID, &customer.Name, &customer.DocumentType, &customer.DocumentNumber, &customer.TaxCondition,
		&customer.Email, &customer.Phone, &customer.Address, &customer.Notes, &customer.System,
		&customer.CreatedAt, &customer.UpdatedAt)

	return customer, err
}

func customerMatches(customer domain.Customer, search, searchDigits string) bool {
	if search == "" {
		return true
	}
	if strings.Contains(domain.NormalizeText(customer.Name), search) {
		return true
	}
	if searchDigits != "" && strings.Contains(customer.DocumentNumber, searchDigits) {
		return true
	}
	for _, contact := range []*string{customer.Email, customer.Phone} {
		if contact != nil && strings.Contains(domain.NormalizeText(*contact), search) {
			return true
		}
	}

	return false
}

func keepDigit(r rune) rune {
	if unicode.IsDigit(r) {
		return r
	}
	return -1
}

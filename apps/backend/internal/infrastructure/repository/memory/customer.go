package memory

import (
	"cmp"
	"context"
	"slices"
	"strings"
	"unicode"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

var _ gateway.CustomerRepository = (*CustomerRepository)(nil)

type CustomerRepository struct {
	store *Store
}

func NewCustomerRepository(store *Store) *CustomerRepository {
	return &CustomerRepository{store: store}
}

func (repository *CustomerRepository) Create(ctx context.Context, customer domain.Customer) (domain.Customer, error) {
	var created domain.Customer
	err := repository.store.run(ctx, func(current *tx) error {
		customer.ID = newID()
		if err := repository.checkDocumentIsFree(customer); err != nil {
			return err
		}
		put(current, repository.store.customers, customer.ID, cloneCustomer(customer))
		created = cloneCustomer(customer)
		return nil
	})

	return created, err
}

func (repository *CustomerRepository) Get(ctx context.Context, id string) (domain.Customer, error) {
	var found domain.Customer
	err := repository.store.run(ctx, func(*tx) error {
		customer, ok := repository.store.customers[id]
		if !ok {
			return domain.ErrCustomerNotFound
		}
		found = cloneCustomer(customer)
		return nil
	})

	return found, err
}

func (repository *CustomerRepository) Update(ctx context.Context, customer domain.Customer) (domain.Customer, error) {
	var updated domain.Customer
	err := repository.store.run(ctx, func(current *tx) error {
		if _, ok := repository.store.customers[customer.ID]; !ok {
			return domain.ErrCustomerNotFound
		}
		if err := repository.checkDocumentIsFree(customer); err != nil {
			return err
		}
		put(current, repository.store.customers, customer.ID, cloneCustomer(customer))
		updated = cloneCustomer(customer)
		return nil
	})

	return updated, err
}

func (repository *CustomerRepository) Delete(ctx context.Context, id string) error {
	return repository.store.run(ctx, func(current *tx) error {
		if _, ok := repository.store.customers[id]; !ok {
			return domain.ErrCustomerNotFound
		}
		if repository.store.hasSaleWhere(func(sale domain.Sale) bool { return sale.Customer.ID == id }) {
			return domain.ErrCustomerInUse
		}
		remove(current, repository.store.customers, id)
		return nil
	})
}

func (repository *CustomerRepository) List(ctx context.Context, filter gateway.CustomerFilter) (domain.Page[domain.Customer], error) {
	page := domain.Page[domain.Customer]{Items: []domain.Customer{}, Page: filter.Page.Page, Limit: filter.Page.Limit}
	err := repository.store.run(ctx, func(*tx) error {
		search := domain.NormalizeText(filter.Search)
		searchDigits := strings.Map(keepDigit, filter.Search)
		var matches []domain.Customer
		for _, customer := range repository.store.customers {
			if customerMatches(customer, search, searchDigits) {
				matches = append(matches, customer)
			}
		}
		slices.SortFunc(matches, func(a, b domain.Customer) int {
			return cmp.Or(cmp.Compare(domain.NormalizeText(a.Name), domain.NormalizeText(b.Name)), cmp.Compare(a.ID, b.ID))
		})

		page.Total = len(matches)
		start := min(filter.Page.Offset(), len(matches))
		end := min(start+filter.Page.Limit, len(matches))
		for _, customer := range matches[start:end] {
			page.Items = append(page.Items, cloneCustomer(customer))
		}
		return nil
	})

	return page, err
}

// checkDocumentIsFree enforces what a partial unique index on the document
// type and number (when there is one) enforces in PostgreSQL.
func (repository *CustomerRepository) checkDocumentIsFree(customer domain.Customer) error {
	key := customer.DocumentKey()
	if key == "" {
		return nil
	}
	for id, other := range repository.store.customers {
		if id != customer.ID && other.DocumentKey() == key {
			return domain.ErrCustomerDocumentInUse
		}
	}

	return nil
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

func cloneCustomer(customer domain.Customer) domain.Customer {
	customer.Email = cloneString(customer.Email)
	customer.Phone = cloneString(customer.Phone)
	customer.Address = cloneString(customer.Address)
	customer.Notes = cloneString(customer.Notes)

	return customer
}
